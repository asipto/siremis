package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

const siregisVersion = "1.0.0"

// CLIOptions - structure for command line options
type GMCLIOptions struct {
	config      string
	domain      string
	httpsrv     string
	httpssrv    string
	httpsusele  bool
	httpspubkey string
	httpsprvkey string
	version     bool
}

var GMCLIOptionsV = GMCLIOptions{
	config:      "etc/config.json",
	domain:      "",
	httpsrv:     ":8284",
	httpssrv:    "",
	httpsusele:  false,
	httpspubkey: "",
	httpsprvkey: "",
	version:     false,
}

type GMAlert struct {
	Active bool
	Type   string
	Text   string
}

type GMResultAttrs struct {
	NrRows      int
	NrGroup     int
	NrGroupPrev int
	NrGroupNext int
}

type GMViewContext struct {
	Action       string
	AuthOK       bool
	SchemaName   string
	SchemaTitle  string
	IdField      GMSchemaField
	IdFieldValue any
	ResultAttrs  GMResultAttrs
	Alert        GMAlert
}

type GMViewFormField struct {
	Field        GMSchemaField
	Value        any
	OptionValues []string
}

type GMViewData struct {
	Config     GMConfig
	Schema     GMSchema
	Context    GMViewContext
	Fields     []GMSchemaField
	Values     []any
	FormFields []GMViewFormField
}

type GMDBField struct {
	Name   string
	Column string
	Value  any
}

var GMTemplatesV *template.Template = nil

func startHTTPServices() chan error {

	errchan := make(chan error)

	// starting HTTP server
	if len(GMCLIOptionsV.httpsrv) > 0 {
		go func() {
			if len(GMConfigV.URLDir) > 0 {
				log.Printf("staring HTTP service on: http://%s%s ...",
					GMCLIOptionsV.httpsrv, GMConfigV.URLDir)
			} else {
				log.Printf("staring HTTP service on: http://%s ...", GMCLIOptionsV.httpsrv)
			}

			if err := http.ListenAndServe(GMCLIOptionsV.httpsrv, nil); err != nil {
				errchan <- err
			}

		}()
	}

	// starting HTTPS server
	if len(GMCLIOptionsV.httpssrv) > 0 && len(GMCLIOptionsV.httpspubkey) > 0 && len(GMCLIOptionsV.httpsprvkey) > 0 {
		go func() {
			if len(GMConfigV.URLDir) > 0 {
				log.Printf("Staring HTTPS service on: https://%s%s ...", GMCLIOptionsV.httpssrv, GMConfigV.URLDir)
			} else {
				log.Printf("Staring HTTPS service on: https://%s ...", GMCLIOptionsV.httpssrv)
			}
			if len(GMCLIOptionsV.domain) > 0 {
				dtoken := strings.Split(strings.TrimSpace(GMCLIOptionsV.httpssrv), ":")
				if len(GMConfigV.URLDir) > 0 {
					log.Printf("HTTPS with domain: https://%s:%s%s ...", GMCLIOptionsV.domain, dtoken[1], GMConfigV.URLDir)
				} else {
					log.Printf("HTTPS with domain: https://%s:%s ...", GMCLIOptionsV.domain, dtoken[1])
				}
			}
			if err := http.ListenAndServeTLS(GMCLIOptionsV.httpssrv, GMCLIOptionsV.httpspubkey, GMCLIOptionsV.httpsprvkey, nil); err != nil {
				errchan <- err
			}
		}()
	}

	return errchan
}

func printCLIOptions() {
	type CLIOptionDef struct {
		Ops      []string
		Usage    string
		DefValue string
		VType    string
	}
	var items []CLIOptionDef
	flag.VisitAll(func(f *flag.Flag) {
		var found bool = false
		for idx, it := range items {
			if it.Usage == f.Usage {
				found = true
				it.Ops = append(it.Ops, f.Name)
				items[idx] = it
			}
		}
		if !found {
			items = append(items, CLIOptionDef{
				Ops:      []string{f.Name},
				Usage:    f.Usage,
				DefValue: f.DefValue,
				VType:    fmt.Sprintf("%T", f.Value),
			})
		}
	})
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Ops[0]) <
			strings.ToLower(items[j].Ops[0])
	})
	for _, val := range items {
		vtype := val.VType[6 : len(val.VType)-5]
		if vtype[len(vtype)-2:] == "64" {
			vtype = vtype[:len(vtype)-2]
		}
		for _, opt := range val.Ops {
			if vtype == "bool" {
				fmt.Printf("  -%s\n", opt)
			} else {
				fmt.Printf("  -%s %s\n", opt, vtype)
			}
		}
		if vtype != "bool" && len(val.DefValue) > 0 {
			fmt.Printf("      %s [default: %s]\n", val.Usage, val.DefValue)
		} else {
			fmt.Printf("      %s\n", val.Usage)
		}
	}
}

// initialize application components
func init() {
	// command line arguments
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s (v%s):\n", filepath.Base(os.Args[0]), siregisVersion)
		printCLIOptions()
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}

	flag.StringVar(&GMCLIOptionsV.config, "config", GMCLIOptionsV.config, "path to json config file")
	flag.StringVar(&GMCLIOptionsV.domain, "domain", GMCLIOptionsV.domain, "http service domain")
	flag.StringVar(&GMCLIOptionsV.httpsrv, "http-srv", GMCLIOptionsV.httpsrv, "http server bind address")
	flag.StringVar(&GMCLIOptionsV.httpssrv, "https-srv", GMCLIOptionsV.httpssrv, "https server bind address")
	flag.StringVar(&GMCLIOptionsV.httpspubkey, "https-pubkey", GMCLIOptionsV.httpspubkey, "https server public key")
	flag.StringVar(&GMCLIOptionsV.httpsprvkey, "https-prvkey", GMCLIOptionsV.httpsprvkey, "https server private key")
	flag.BoolVar(&GMCLIOptionsV.httpsusele, "use-letsencrypt", GMCLIOptionsV.httpsusele,
		"use local letsencrypt certificates (requires domain)")
	flag.BoolVar(&GMCLIOptionsV.version, "version", GMCLIOptionsV.version, "print version")
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if GMCLIOptionsV.httpsusele && len(GMCLIOptionsV.domain) == 0 {
		log.Printf("use-letsencrypt requires domain parameter\n")
		os.Exit(1)
	}

	configBytes, err := os.ReadFile(GMCLIOptionsV.config)
	if err != nil {
		log.Printf("unavailable config file %s\n", GMCLIOptionsV.config)
		os.Exit(1)
	}
	err = json.Unmarshal(configBytes, &GMConfigV)
	if err != nil {
		log.Printf("invalid content in config file %s\n", GMCLIOptionsV.config)
		os.Exit(1)
	}

	GMConfigEvalVals()

	if _, err := os.Stat(GMConfigV.PublicDir); os.IsNotExist(err) {
		log.Printf("%s folder cannot be found\n", GMConfigV.PublicDir)
		os.Exit(1)
	}

	if GMCLIOptionsV.httpsusele && len(GMCLIOptionsV.httpssrv) > 0 && len(GMCLIOptionsV.domain) > 0 {
		GMCLIOptionsV.httpspubkey = "/etc/letsencrypt/live/" + GMCLIOptionsV.domain + "/fullchain.pem"
		GMCLIOptionsV.httpsprvkey = "/etc/letsencrypt/live/" + GMCLIOptionsV.domain + "/privkey.pem"
	}

	GMFuncMap["HA1"] = GMFuncHA1
	GMFuncMap["HA1B"] = GMFuncHA1B
	GMFuncMap["DateTimeNow"] = GMFuncDateTimeNow
	GMFuncMap["DBColumnValues"] = GMFuncDBColumnValues

	GMTemplatesV = template.Must(template.New("").Funcs(template.FuncMap{
		"rowon":     GMTemplateFuncRowOn,
		"add":       GMTemplateFuncAdd,
		"sub":       GMTemplateFuncSub,
		"mod":       GMTemplateFuncMod,
		"loop":      GMTemplateFuncLoop,
		"lastloop":  GMTemplateFuncLastLoop,
		"lastindex": GMTemplateFuncLastIndex,
	}).ParseGlob("templates/*"))

	http.Handle(GMConfigV.URLDir+"/"+GMConfigV.PublicDir+"/",
		http.StripPrefix(strings.TrimRight(GMConfigV.URLDir+"/"+GMConfigV.PublicDir+"/", "/"),
			http.FileServer(http.Dir(GMConfigV.URLDir+"/"+GMConfigV.PublicDir))))
	http.HandleFunc("/", GMRequestHandler)
	// http.HandleFunc("/show", Show)
	// http.ListenAndServe(":8284", nil)
	errchan := startHTTPServices()
	errx := <-errchan
	log.Printf("unable to start http services due to (error: %v)", errx)
	os.Exit(1)
}
