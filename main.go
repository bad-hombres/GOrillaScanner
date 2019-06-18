package main

import (
  "flag"
  "fmt"
  "os"
  "strings"
  "strconv"
  "GOrillaScanner/scanner"
  "GOrillaScanner/log"
  "GOrillaScanner/helpers"
  "time"
)

func parseList(value string) (list []int64) {
  if len(value) == 0 {
    return
  }

  for _, c := range strings.Split(value, ",") {
    n, err := strconv.ParseInt(c, 10, 64)
    if err != nil {
      log.Error("Need to pass valid numbers")
      os.Exit(1)
    }
    
    list = append(list, n)
  }

  return
}

func parseLengths(value string) (list []string) {
  if len(value) == 0 {
    return
  }

  for _, c := range strings.Split(value, ",") {
    list = append(list, c)
  }

  return
}

func main() {
  helpers.PrintBanner()

  url := flag.String("u", "", "Url to scan")
  wordlist := flag.String("w", "", "Wordlist to use to scan")
  threads := flag.Int("t", 5, "Threads to use")
  tmpcodes := flag.String("codes", "", "Comma seperated list of codes to use")
  tmplen := flag.String("lengths", "", "Comma seperated list of contentLengths to use")
  hide := flag.Bool("hide", false, "Used in conjuction with codes to hide those codes instead of printing")
  timeout := flag.String("timeout", "2s", "Timeout for HTTP requests")
  flag.Parse()

  if *url == "" || !strings.Contains(*url, "FUZZ") {
    log.Error("Must provide -u option and it must contain FUZZ")
    os.Exit(1)
  }

  if _, err := os.Stat(*wordlist); *wordlist == "" || os.IsNotExist(err) {
    log.Error("Must provide valid wordlist file")
    os.Exit(1)
  }

  if *tmplen == "" && *tmpcodes == "" {
    log.Error("Need to provide codes or lengths")
    os.Exit(1)
  }
  
  duration, err := time.ParseDuration(*timeout)
  if err != nil {
    log.Error("Invalid timeout passed")
    os.Exit(1)
  }

  codes := parseList(*tmpcodes)
  lengths := parseLengths(*tmplen)

  log.Info(fmt.Sprintf("Parsed options using url: %s, wordlist: %s", *url, *wordlist))
  s := scanner.New(*wordlist, *threads, duration)
  s.Start(*url, codes, lengths, *hide)
}
