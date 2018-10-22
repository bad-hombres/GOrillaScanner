package scanner

import (
  "fmt"
)

type Result struct {
  url string
  code int
  contentlen int64
}

func (r *Result) String() string {
  return fmt.Sprintf("%s -> {Code: %d, Content Len: %d}", r.url, r.code, r.contentlen)
}
