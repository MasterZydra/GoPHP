package main

import (
	"GoPHP/cmd/goPHP/common"
	"GoPHP/cmd/goPHP/config"
	"GoPHP/cmd/goPHP/ini"
	"GoPHP/cmd/goPHP/interpreter"
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var serverAddr string
var documentRoot string
var isDevMode bool

func main() {
	file := flag.String("f", "", "Parse and execute <file>.")
	isDev := flag.Bool("dev", false, "Run in developer mode.")
	// Web server
	addr := flag.String("S", "", "Run with built-in web server. <addr>:<port>")
	docRoot := flag.String("t", "", "Specify document root <docroot> for built-in web server.")

	flag.Parse()

	isDevMode = *isDev

	// Serve with built-in web server
	if *addr != "" {
		serverAddr = *addr
		documentRoot = *docRoot
		webServer()
		os.Exit(0)
	}

	// Parse given file
	if *file != "" {
		absFilePath, err := common.GetAbsPath(*file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		processFile(absFilePath)
	}

	// Read stdin or wait for it
	processStdin()
}

// ------------------- MARK: stdin -------------------

func processStdin() {
	content := ""
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if content != "" || scanner.Text() == "" {
			content += "\n"
		}
		content += scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}

	output, exitCode := processContent(nil, string(content), "main.php")
	fmt.Print(output)
	os.Exit(exitCode)
}

// ------------------- MARK: given file -------------------

func processFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	output, exitCode := processContent(nil, string(content), filename)
	fmt.Print(output)
	if exitCode == 500 {
		exitCode = 1
	}
	os.Exit(exitCode)
}

// ------------------- MARK: web server -------------------

func webServer() {
	if documentRoot == "" {
		documentRoot, _ = os.Getwd()
	} else {
		absPath, err := common.GetAbsPath(documentRoot)
		if err != nil {
			fmt.Println("Error: Could not find directory " + documentRoot)
			os.Exit(1)
		}
		documentRoot = absPath
	}

	var mode string
	if isDevMode {
		mode = "Development"
	} else {
		mode = "Production"
	}

	fmt.Printf("[%s] GoPHP %s %s Server (%s) started\n",
		time.Now().Format("Mon Jan 02 15:04:05 2006"), config.Version, mode, serverAddr,
	)
	fmt.Println("Document root is " + documentRoot)
	fmt.Println("Press Ctrl-C to quit")

	http.HandleFunc("/", requestHandler)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		fmt.Println("Failed to start web server")
		fmt.Println(err)
	}
}

func getNotFoundText(resource string) string {
	return `<!doctype html><html><head><title>404 Not Found</title><style>
			body { background-color: #fcfcfc; color: #333333; margin: 0; padding:0; }
			h1 { font-size: 1.5em; font-weight: normal; background-color: #9999cc; min-height:2em; line-height:2em; border-bottom: 1px inset black; margin: 0; }
			h1, p { padding-left: 10px; }
			code.url { background-color: #eeeeee; font-family:monospace; padding:0 2px;}
			</style>
			</head><body><h1>Not Found</h1><p>The requested resource <code class="url">` + resource + `</code> was not found on this server.</p></body></html>
		`
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	absFilePath := path.Join(documentRoot, r.URL.Path)
	_, err := os.Stat(absFilePath)
	if err != nil {
		if isDevMode {
			fmt.Println("404", absFilePath)
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, getNotFoundText(r.URL.Path))
		return
	}

	if common.IsDir(absFilePath) {
		index := path.Join(absFilePath, "index.html")
		if common.PathExists(index) && !common.IsDir(index) {
			absFilePath = index
		} else if index = path.Join(absFilePath, "index.php"); common.PathExists(index) && !common.IsDir(index) {
			absFilePath = index
		} else {
			if isDevMode {
				fmt.Println("404", absFilePath)
			}
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, getNotFoundText(r.URL.Path))
			return
		}
	}

	if !strings.HasSuffix(absFilePath, ".php") {
		http.ServeFile(w, r, absFilePath)
		return
	}

	content, err := os.ReadFile(absFilePath)
	if err != nil {
		fmt.Println("Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO content-type returned from interpreter?
	w.Header().Set("Content-Type", "text/html")
	output, exitCode := processContent(r, string(content), absFilePath)
	if exitCode == 500 {
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprint(w, output)
}

// ------------------- MARK: core logic -------------------

func processContent(r *http.Request, content string, filename string) (output string, exitCode int) {
	var initIni *ini.Ini
	if isDevMode {
		initIni = ini.NewDevIni()
	} else {
		initIni = ini.NewDefaultIni()
	}

	request := interpreter.NewRequestFromGoRequest(r, documentRoot, serverAddr, filename)
	interpreter := interpreter.NewInterpreter(initIni, request, filename)
	result, err := interpreter.Process(content)
	if err != nil {
		result += interpreter.ErrorToString(err)
		return result, 500
	}
	return result, interpreter.GetExitCode()
}
