package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"strings"

	"log"

	"flag"

	"goftp.io/server/v2"
	"goftp.io/server/v2/driver/file"
)

func fileServerWithDetails(root http.FileSystem) http.Handler {
    fileTypesSameEmoji := map[string][]string{
        "📦": {"zip", "gzip", "gz", "7z", "rar", "tar", "xz", "bz2"},
        "🗒️": {"txt", "text"},
        "🖼️": {"jpg", "jpeg", "png", "gif", "bmp", "svg", "ico", "webp", "tiff", "tif"},
        "📕": {"doc", "docx", "pdf"},
        "💾": {"config", "cfg", "exe", "bat", "sh", "go", "html", "css", "js", "json", "java", "class", "jar", "jsp", "jspx", "py", "pyc", "pyd", "pyo", "pyw", "pyx", "pyz", "c", "h", "cpp", "hpp", "cc", "cxx", "hxx", "ino", "cs", "csx", "swift", "ipa", "dart", "asm", "r", "rmd", "sql", "asp", "aspx", "axd", "asx", "asmx", "ashx", "json", "ps1", "psm1", "psd1", "ps1xml", "pssc", "psc1", "js", "mjs", "ts", "tsx"},
        "📄": {"xml", "csv", "xlsx", "xls", "tsv"},
        "📜": {"log"},
        "🗃️": {"folder"},
        "🎥": {"mp4", "mkv", "avi", "mov", "wmv", "flv", "webm", "m4v", "3gp", "mpeg"},
        "🔊": {"mp3", "wav", "ogg", "flac", "aac", "wma", "m4a", "amr", "aiff", "opus"},
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        filePath := filepath.Join(".", r.URL.Path)
        file, err := root.Open(filePath)
        if err != nil {
            http.NotFound(w, r)
            return
        }
        defer file.Close()

        stat, err := file.Stat()
        if err != nil {
            http.Error(w, fmt.Sprintf("Could not get file info: %s", err), http.StatusInternalServerError)
            return
        }

        if stat.IsDir() {
            // Display directory listing
            files, err := file.Readdir(-1)
            if err != nil {
                http.Error(w, fmt.Sprintf("Could not read directory: %s", err), http.StatusInternalServerError)
                return
            }

            w.Header().Set("Content-Type", "text/html")


            fmt.Fprintf(w, "<!DOCTYPE html>")
            fmt.Fprintf(w, "<html>")
            fmt.Fprintf(w, "<head>")
            fmt.Fprintf(w, "<meta charset=\"UTF-8\">")
            fmt.Fprintf(w, "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">")
            fmt.Fprintf(w, "<title>Directory listing for %s</title>", r.URL.Path)
            fmt.Fprintf(w, "<style>")
	    fmt.Fprintf(w, "body { background-color:  #121212;font-family: Arial, sans-serif; }")
            fmt.Fprintf(w, "ul { list-style-type: none; padding: 0; }")
	    fmt.Fprintf(w, "li { margin-bottom: 5px; color: white; }")
            fmt.Fprintf(w, "a { color: #df62a0;}")

	    fmt.Fprintf(w, "h2 { color: white;}")
            fmt.Fprintf(w, "</style>")
            fmt.Fprintf(w, "</head>")

	    fmt.Fprintf(w, "<h2>Directory listing for %s</h2>", r.URL.Path)
            fmt.Fprintln(w, "<ul>")

for _, fileInfo := range files {
                fileName := fileInfo.Name()
                if fileInfo.IsDir() {
                    // Use folder emoji for directories
                    fileName += "/"
                    fileTypeEmoji := "🗂️"
		    fmt.Fprintf(w, "<li>%s <a href=\"%s\">%s</a></li>", fileTypeEmoji, fileName, fileName)

                } else {
                    // Get emoji for file type
                    var fileTypeEmoji string
                    ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
                    for emoji, types := range fileTypesSameEmoji {
                        for _, t := range types {
                            if t == ext {
                                fileTypeEmoji = emoji
                                break
                            }
                        }
                        if fileTypeEmoji != "" {
                            break
                        }
                    }
                    if fileTypeEmoji == "" {
                        fileTypeEmoji = "📄" // Default emoji
                    }

                    // Create a link for the file
                    fmt.Fprintf(w, "<li>%s <a href=\"%s\">%s</a> - %d B</li>", fileTypeEmoji, fileName, fileName, fileInfo.Size())
                }
            }
            fmt.Fprintln(w, "</ul>")
            fmt.Fprintf(w, "</body>")
            fmt.Fprintf(w, "</html>")
        } else {
            // Serve file
            http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
        }
    })
}


var (
    rootDir   = flag.String("rootDir", "./", "Set the root directory for serving files.")
    httpPort  = flag.Int("httpPort", 8080, "Set the port for the HTTP file browser.")
    httpAddr  = flag.String("httpAddr", "0.0.0.0", "Set the hostname for the HTTP file browser.")
    ftpPort   = flag.Int("ftpPort", 2000, "Set the port for the FTP server.")
    ftpAddr   = flag.String("ftpAddr", "0.0.0.0", "Set the hostname for the FTP server.")
    userName  = flag.String("ftpUser", "root", "Set the FTP username.")
    ftpPasswd = flag.String("ftpPasswd", "space", "Set the FTP server password.")
)



func main() {
	flag.Parse()
	http.Handle("/", fileServerWithDetails(http.Dir(*rootDir)))

	driver, err := file.NewDriver(*rootDir)
	if err != nil {
		log.Fatal(err)
	}

	ftp, err := server.NewServer(&server.Options{
		Driver:   driver,
		Hostname: *ftpAddr,
		Port:     *ftpPort,
		Auth: &server.SimpleAuth{
			Name:     *userName,
			Password: *ftpPasswd,
		},
		Perm: server.NewSimplePerm("root", "root"),
	})
	if err != nil {
		log.Fatal(err)
	}





        go func() {
                addr := fmt.Sprintf("%s:%d", *httpAddr, *httpPort)
                log.Printf("HTTP Listening on %s\n", addr)
                if err := http.ListenAndServe(addr, logRequest(http.DefaultServeMux)); err != nil {
                        log.Printf("Error starting HTTP server: %s\n", err)
                }
        }()


	log.Printf("Start FTP server on %s:%d", *ftpAddr, *ftpPort)
	if err := ftp.ListenAndServe(); err != nil {
		log.Fatal(err)
	}




}


func logRequest(handler http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
                handler.ServeHTTP(w, r)
        })
}
