package request

import (
	"fmt"
	"testing"
)

func TestHttp(t *testing.T) {
	http := Request{BaseUri: "https://hao.360.com/"}
	fmt.Println(http.Get(""))
}
