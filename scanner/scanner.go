package scanner

import (
  "GOrillaScanner/log"
  "fmt"
  "strings"
  "net/http"
  "bufio"
  "os"
  "time"
  "io/ioutil"
  "strconv"
)

type Scanner struct {
  wordlist string
  results chan *Result
  client *http.Client
  threads int
  scheduler *Scheduler
}


func (s *Scanner) printDebugResponse(r *Result, count int) {
    fmt.Printf("\x1b[2K")
    log.Debug(fmt.Sprintf("[ threads: %v, Responses: %v] - %v", s.scheduler.GetThreadCount(), count, r.String()))
    fmt.Printf("\x1b[1A")
}

func (s *Scanner) printFoundResponse(r *Result) {
  fmt.Printf("\x1b[2K")
  log.Info(r.String())
}

func (s *Scanner) recieveResults(filter func(resp *Result) bool) {
  results := 0
  for r := range s.results {
      results++
      s.printDebugResponse(r, results)
      if filter(r) {
        s.printFoundResponse(r)
      } 
  }
}

func (s *Scanner) makeRequest(url string) {
  resp, err := s.client.Get(url)
  if err != nil {
    s.results <- &Result{url, -1, 0}
    return
  }

  contentLen := resp.ContentLength
  if (contentLen == -1) {
    body, _ := ioutil.ReadAll(resp.Body)
    contentLen = int64(len(body))
  }

  defer resp.Body.Close()
  s.results <- &Result{url, resp.StatusCode, contentLen}
}

func checkVals(val int64, vals []int64) bool {
    if len(vals) == 0 {
      return true
    }

    for _, v := range vals {
      if val == v {
        return true
      }
    }
    return false
}

func checkLengths(val int64, vals []string) bool {
  op := "="
  var tmpVal int64
  var tmpVals [2]int64

  for _, v := range vals {
    if (strings.Contains(v, ">")) {
      tmp := strings.Split(v, ">")[1]
      op = ">"
      tmpVal, _ = strconv.ParseInt(tmp, 10, 64)
    }

    if (strings.Contains(v, "<")) {
      tmp := strings.Split(v, "<")[1]
      op = "<"
      tmpVal, _ = strconv.ParseInt(tmp, 10, 64)
    }

    if (strings.Contains(v, "..")) {
      tmp := strings.Split(v, "..")
      op = ".."
      tmpVals[0], _ = strconv.ParseInt(tmp[0], 10, 64)
      tmpVals[1], _ = strconv.ParseInt(tmp[1], 10, 64)
    }

    if (strings.Contains(v, "=")) {
      tmp := strings.Split(v, "=")[1]
      op = "="
      tmpVal, _ = strconv.ParseInt(tmp, 10, 64)
    }

    switch op {
    case ">":
      if (val > tmpVal) {
        return true
      }
    case "<":
      if (val < tmpVal) {
        return true
      }
    case "..":
      //log.Warn(fmt.Sprintf("%v - %v = %v", tmpVals, val, val >= tmpVals[0] && val <= tmpVals[1]))
      if (val >= tmpVals[0] && val <= tmpVals[1]) {
        return true
      }
    default:
      if (val == tmpVal) {
        return true
      }
    }
  }
  return false;
}

func (s *Scanner) Start(template string, codes []int64, lengths []string, hide bool) {
  log.Warn(fmt.Sprintf("Starting scan of url: %s", template))
  s.results = make(chan *Result)

  go s.recieveResults(func(resp *Result) bool {
      if hide {
        return !(checkVals(int64(resp.code), codes) && checkLengths(resp.contentlen, lengths))
      } else {
        return checkVals(int64(resp.code), codes) && checkLengths(resp.contentlen, lengths)
      }
  })
  
  file, err := os.Open(s.wordlist)
  if err != nil {
    log.Error("Could not open wordlist")
    return
  }

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
     word := scanner.Text() 
     s.scheduler.StartThread(func() {
       s.makeRequest(strings.Replace(template, "FUZZ", word, -1))
     })
  }
  s.scheduler.WaitForThreads(true)
  close(s.results)
}

func New(wordlist string, threads int, timeout time.Duration) *Scanner {
  return &Scanner{
    wordlist: wordlist,
    client: &http.Client{
      CheckRedirect: func(req *http.Request, via []*http.Request) error { 
                            return http.ErrUseLastResponse 
                     }, 
      Timeout: timeout}, 
    scheduler: NewScheduler(threads)}
}
