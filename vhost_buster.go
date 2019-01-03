package main

import (
	"io"
	"fmt"
	"time"
	"sync"
	"regexp"
	"flag"
  "bufio"
  "os"
	"log"
  "strconv"
	"strings"
	"net/http"
	"github.com/imroc/req"
 	"io/ioutil"
  "crypto/tls"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

//Declaring Number of Threads
var THREADS int64
var DEBUG bool

//Cleaing the tokens, newlines, spaces and all unwanted numbers etc, so that we can compare the response efficiently
func cleanStr(bodyString string) string {
	//Removing all new lines
	re := regexp.MustCompile(`\r?\n`)
	bodyString = re.ReplaceAllString(bodyString, "")

	//Removing all hex tokens
	re = regexp.MustCompile(`[a-fA-F0-9]{10,}`)
	bodyString = re.ReplaceAllString(bodyString, "")

	//Removing all numbers
	re = regexp.MustCompile(`\d+`)
	bodyString = re.ReplaceAllString(bodyString, "")
	return bodyString
}

//Function to check and log errors
func check(e error){
	if e != nil {
        	log.Fatal(e)
    	}
}

func main() {
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	ScanText := ""
	var wg_threads sync.WaitGroup
	banner := `                                                                            ;                                             
                                                                            ED.                   :                       
          .        ,;      ., :                                             E#Wi                 t#,                     .
         ;W      f#i      ,Wt Ef      j.         t                      t   E###G.       t      ;##W.                   ;W
        f#E    .E#t      i#D. E#t     EW,        Ej GEEEEEEELf.     ;WE.Ej  E#fD#W;      Ej    :#L:WE  GEEEEEEEL       f#E
      .E#f    i#W,      f#f   E#t     E##j       E#,,;;L#K;;.E#,   i#G  E#, E#t t##L     E#,  .KG  ,#D ,;;L#K;;.     .E#f 
     iWW;    L#D.     .D#i    E#t     E###D.     E#t   t#E   E#t  f#f   E#t E#t  .E#K,   E#t  EE    ;#f   t#E       iWW;  
    L##Lffi:K#Wfff;  :KW,     E#t fi  E#jG#W;    E#t   t#E   E#t G#i    E#t E#t    j##f  E#t f#.     t#i  t#E      L##Lffi
   tLLG##L i##WLLLLt t#f      E#t L#j E#t t##f   E#t   t#E   E#jEW,     E#t E#t    :E#K: E#t :#G     GK   t#E     tLLG##L 
     ,W#i   .E#L      ;#G     E#t L#L E#t  :K#E: E#t   t#E   E##E.      E#t E#t   t##L   E#t  ;#L   LW.   t#E       ,W#i  
    j#E.      f#E:     :KE.   E#tf#E: E#KDDDD###iE#t   t#E   E#G        E#t E#t .D#W;    E#t   t#f f#:    t#E      j#E.   
  .D#j         ,WW;     .DW:  E###f   E#f,t#Wi,,,E#t   t#E   E#t        E#t E#tiW#G.     E#t    f#D#;     t#E    .D#j     
 ,WK,           .D#;      L#, E#K,    E#t  ;#W:  E#t   t#E   E#t        E#t E#K##i       E#t     G#t      t#E   ,WK,      
 EG.              tt       jt EL      DWi   ,KK: E#t    fE   EE.        E#t E##D.        E#t      t        fE   EG.       
 ,                            :                  ,;.     :   t          ,;. E#t          ,;.                :   ,         
                                                                            L:                                            
`
	fmt.Printf("%s",banner)
	
  //Reading command line args
	hostlistPtr := flag.String("host", "hostlist.txt", "File Path for host names list")
	iplistPtr := flag.String("ip", "iplist.txt", "Ip Addresses list for the server to test")
	tPtr := flag.Int("t", 20, "Number of threads")
	debugPtr := flag.Bool("d", false, "Set -d Flag to turn on debugging")
	//Will think if they are usefull to be implemented in future
	//ipPtr := flag.String("ip", "127.0.0.1", "Ip Address for the server to test")
    	//ssltestPtr := flag.Bool("ssltest", false, "HTTPS Test On/Off flag")
	flag.Parse()


	fmt.Println("HostFile:", *hostlistPtr)
	fmt.Println("IP List:", *iplistPtr)
	fmt.Println("Threads:", *tPtr)
	THREADS = int64(*tPtr)
	DEBUG = *debugPtr

	//Reading Ip List file
	file, err := os.Open(*iplistPtr)
    	check(err)
    	defer file.Close()
	total := countHosts(hostlistPtr)+1
	fmt.Println("Host Count:"+strconv.Itoa(total)+"\n")

    	scanner := bufio.NewScanner(file)
    	for scanner.Scan() {
		ScanText = scanner.Text()
		wg.Add(total)
		go checkHosts(&wg,&wg_threads,ScanText,hostlistPtr,total,p)
    	}
	check(scanner.Err())
	// wait for all bars to complete and flush
	p.Wait()
}

//Function to enumerate the host file and call triggerReq to function for each host.
func checkHosts(wg *sync.WaitGroup,wg_threads *sync.WaitGroup,IP string,filePtr *string,total int,p *mpb.Progress) {
	
	j := int64(0)
	output := ""
	name := IP
	startTime := time.Now()

	//Creating progressbar for each IP to be checked
	efn := func(w io.Writer, s *decor.Statistics) {
		fmt.Fprintf(w,output)
	}
	bar := p.AddBar(int64(total), mpb.BarNewLineExtend(efn),
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name(name),
			// decor.DSyncWidth bit enables column width synchronization
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 60
				decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
			),
		),
	)
  //Getting the default response code and output length from server.
	wg_threads.Add(1)
	invalidStatus,invalidLength := triggerReq(wg,wg_threads,IP,"infinityw0rri0rt3st",0,0,bar,startTime,&output)

	//Reading Host file
	hostfile, err := os.Open(*filePtr)
	check(err)

	hostscanner := bufio.NewScanner(hostfile)
	for hostscanner.Scan() {
		j = state(wg_threads)
		for j > THREADS{
			time.Sleep(100 * time.Millisecond)
			j = state(wg_threads)
		}
		wg_threads.Add(1)
    //Calling trigger request function to trigger the request with the specified host header.
		go triggerReq(wg,wg_threads,IP,hostscanner.Text(),invalidStatus,invalidLength,bar,startTime,&output)
	}
	hostfile.Close()
}

//Triggers request and check if it meets the default response. Do not print if meets the default response.
func triggerReq(wg *sync.WaitGroup,wg_threads *sync.WaitGroup,IP string, host string,invalidStatus int,invalidLength int,bar *mpb.Bar,startTime time.Time,output *string) (int,int){
	defer wg_threads.Done()
	defer wg.Done()
	defer bar.IncrBy(1, time.Since(startTime))
	tr := &http.Transport{
        	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    	}
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
        	return http.ErrUseLastResponse
    		},
		Transport: tr}
	header := req.Header{"Host":host}
	if DEBUG{
		req.Debug = true
	}
	req.SetClient(client)
	r, err := req.Get(IP, header)
	//fmt.Printf("Testing : Ip : "+ IP +" Host : "+ host+"\n")
	if err != nil{
		return 0,0
	}
	resp := r.Response()
	defer resp.Body.Close()
    	bodyBytes, _ := ioutil.ReadAll(resp.Body)
    	bodyString := string(bodyBytes)
    	bodyString = cleanStr(bodyString)
	
	//Output
	if invalidStatus==0 || (invalidStatus != resp.StatusCode && invalidLength != len(bodyString)){
		*output = *output + "Ip : "+ IP +" Host : "+ host +" Code : " + strconv.Itoa(resp.StatusCode) + " Length : " + strconv.Itoa(len(bodyString))+"\n"
	}
	return resp.StatusCode,len(bodyString)
}

//Reading the number of threads running to control parrallel threading.
func state(wg *sync.WaitGroup) (int64){
	outp := strings.Split(fmt.Sprintf("%+v\n", wg)," ")
	j, _ := strconv.ParseInt(outp[5], 10, 64)
	k, _ := strconv.ParseInt(outp[6], 10, 64)
	//fmt.Printf("Waiting for : "+ strconv.FormatInt((j + (k * 255)),10)+"\n")
	return (j + (k * 255))
}

//Keeping the file lines count code in a different file.
func countHosts(filePtr *string) (int) {
	//Reading Host file
	hostfile, err := os.Open(*filePtr)
	check(err)
	defer hostfile.Close()
	j := 0

	hostscanner := bufio.NewScanner(hostfile)
	//Counting the numbers of hosts to test
	for hostscanner.Scan() {
		j += 1
	}
	return j
}
