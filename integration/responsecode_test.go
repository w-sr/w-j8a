package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func TestStatusCode100SentFromProxyWithPutIfExpectedReturns200(t *testing.T) {
	client := &http.Client{}
	serverPort := 8080
	wantDownstreamStatusCode := 200

	jsonData := map[string]string{"firstname": "Simon", "lastname": "Mittag", "rank": "Corporal"}
	jsonValue, _ := json.Marshal(jsonData)
	buf := bytes.NewBuffer(jsonValue)

	url := fmt.Sprintf("http://localhost:%d/mse6/put", serverPort)
	req, _ := http.NewRequest("PUT", url, buf)

	req.Header.Add("Expect", "100-continue")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	gotDownstreamStatusCode := 0
	if err != nil {
		t.Errorf("error connecting to upstream for port %d, /send, cause: %v", serverPort, err)
		return
	} else {
		gotDownstreamStatusCode = resp.StatusCode
	}

	if gotDownstreamStatusCode != wantDownstreamStatusCode {
		t.Errorf("PUT with Expect: 100-continue did not result in OK, want %d, got %d", wantDownstreamStatusCode, gotDownstreamStatusCode)
	}
}

//Test normal responses
func TestStatusCodeOfProxiedResponses200To226(t *testing.T) {
	var wg1 sync.WaitGroup
	for i := 200; i <= 226; i++ {
		wg1.Add(1)
		go performJabbaResponseCodeTest(&wg1, t, i, i, 8080)
	}
	wg1.Wait()
}

func TestStatusCode216OfProxiedResponse(t *testing.T) {
	performOneJabbaResponseCodeTest(t, 216, 216, 8080)
}

//Test redirects are mapped through to the calling user agent
func TestStatusCodeOfProxiedResponses300To308NonRedirected(t *testing.T) {
	var wg1 sync.WaitGroup
	for i := 300; i <= 308; i++ {
		wg1.Add(1)
		go performJabbaResponseCodeTest(&wg1, t, i, i, 8080)
	}
	wg1.Wait()
}

func TestStatusCode300SeriesRedirect(t *testing.T) {
	//we want these to redirect
	locHeader := []int{301, 302, 303, 307, 308}
	for _, h := range locHeader {
		client := &http.Client{}
		serverPort := 8080
		getUpstreamStatusCode := h

		//so they should give us a 200 from subsequent request.
		wantDownstreamStatusCode := 200

		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/mse6/send?code=%d", serverPort, getUpstreamStatusCode))
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
		}

		gotDownstreamStatusCode := 0
		if err != nil {
			t.Errorf("error connecting to upstream for port %d, /send, cause: %v", serverPort, err)
			return
		} else {
			gotDownstreamStatusCode = resp.StatusCode
		}

		if gotDownstreamStatusCode != wantDownstreamStatusCode {
			t.Errorf("bad. port %d, testMethod /send, up code %d, want dwn code %d, but got %d", serverPort,
				getUpstreamStatusCode, wantDownstreamStatusCode, gotDownstreamStatusCode)
		} else {
			t.Logf("normal. port %d, testMethod /send, up code %d, want dwn code %d, got %d", serverPort,
				getUpstreamStatusCode, wantDownstreamStatusCode, gotDownstreamStatusCode)
		}
	}
}

//Test client errors are mapped through to the calling user agent
func TestStatusCodeOfProxiedResponses400To499(t *testing.T) {
	var wg1 sync.WaitGroup
	for i := 400; i <= 499; i++ {
		wg1.Add(1)
		go performJabbaResponseCodeTest(&wg1, t, i, i, 8080)
	}
	wg1.Wait()
}

//Test upstream server errors are re-mapped to 502 bad gateway
func TestStatusCodeOfProxiedResponses500To599(t *testing.T) {
	var wg2 sync.WaitGroup
	for i := 500; i <= 599; i++ {
		wg2.Add(1)
		performJabbaResponseCodeTest(&wg2, t, i, 502, 8080)
	}
	wg2.Wait()
}

func performOneJabbaResponseCodeTest(t *testing.T, getUpstreamStatusCode int, wantDownstreamStatusCode int, serverPort int) {
	var wg sync.WaitGroup
	wg.Add(1)
	performJabbaResponseCodeTest(&wg, t, getUpstreamStatusCode, wantDownstreamStatusCode, serverPort)
	wg.Wait()
}

func performJabbaResponseCodeTest(wg *sync.WaitGroup, t *testing.T, getUpstreamStatusCode int, wantDownstreamStatusCode int, serverPort int) {
	//for multithreaded tests we need to count them all down
	defer wg.Done()

	//test client do not follow redirects mate!
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/mse6/send?code=%d", serverPort, getUpstreamStatusCode))
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	gotDownstreamStatusCode := 0
	if err != nil {
		t.Errorf("error connecting to upstream for port %d, /send, cause: %v", serverPort, err)
		return
	} else {
		gotDownstreamStatusCode = resp.StatusCode
	}

	if gotDownstreamStatusCode != wantDownstreamStatusCode {
		t.Errorf("bad. port %d, testMethod /send, up code %d, want dwn code %d, but got %d", serverPort,
			getUpstreamStatusCode, wantDownstreamStatusCode, gotDownstreamStatusCode)
	} else {
		t.Logf("normal. port %d, testMethod /send, up code %d, want dwn code %d, got %d", serverPort,
			getUpstreamStatusCode, wantDownstreamStatusCode, gotDownstreamStatusCode)
	}
}
