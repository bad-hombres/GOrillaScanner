package scanner

import (
  "GOrillaScanner/log"
  "fmt"
  "strings"
  "net/http"
  "bufio"
  "os"
  "time"
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
  defer resp.Body.Close()
  s.results <- &Result{url, resp.StatusCode, resp.ContentLength}
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

func (s *Scanner) Start(template string, codes []int64, lengths []int64, hide bool) {
  log.Warn(fmt.Sprintf("Starting scan of url: %s", template))
  s.results = make(chan *Result)

  go s.recieveResults(func(resp *Result) bool {
      if hide {
        return !(checkVals(int64(resp.code), codes) && checkVals(resp.contentlen, lengths))
      } else {
        return checkVals(int64(resp.code), codes) && checkVals(resp.contentlen, lengths)
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
