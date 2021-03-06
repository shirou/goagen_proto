package goagen_js

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
)

//NewGenerator returns an initialized instance of a JavaScript Client Generator
func NewGenerator(options ...Option) *Generator {
	g := &Generator{}

	for _, option := range options {
		option(g)
	}

	return g
}

// Generator is the application code generator.
type Generator struct {
	API      *design.APIDefinition // The API definition
	OutDir   string                // Destination directory
	Scheme   string                // Scheme used by JavaScript client
	Host     string                // Host addressed by JavaScript client
	genfiles []string              // Generated files
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var (
		outDir, ver  string
		scheme, host string
	)

	set := flag.NewFlagSet("client", flag.PanicOnError)
	set.StringVar(&outDir, "out", "", "")
	set.String("design", "", "")
	set.StringVar(&scheme, "scheme", "", "")
	set.StringVar(&host, "host", "", "")
	set.StringVar(&ver, "version", "", "")
	set.Parse(os.Args[1:])

	// First check compatibility
	if err := codegen.CheckVersion(ver); err != nil {
		return nil, err
	}

	// Now proceed
	g := &Generator{
		OutDir: outDir,
		Scheme: scheme,
		Host:   host,
		API:    design.Design,
	}

	return g.Generate()
}

// Generate produces the skeleton main.
func (g *Generator) Generate() ([]string, error) {
	var err error
	if g.API == nil {
		return nil, fmt.Errorf("missing API definition, make sure design is properly initialized")
	}

	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	if g.Scheme == "" && len(g.API.Schemes) > 0 {
		g.Scheme = g.API.Schemes[0]
	}
	if g.Scheme == "" {
		g.Scheme = "http"
	}
	if g.Host == "" {
		g.Host = g.API.Host
	}
	if g.Host == "" {
		return nil, fmt.Errorf("missing host value, set it with --host")
	}

	if err := os.MkdirAll(g.OutDir, 0755); err != nil {
		return nil, err
	}

	ps, err := parseResources(g)
	if err != nil {
		return nil, err
	}
	outfile := "api.proto"
	file, err := openFile(filepath.Join(g.OutDir, outfile))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buf := bufio.NewWriter(file)

	tmpl, err := newTemplate(
		headerT,
		serviceT,
		messageT,
	)
	if err != nil {
		return nil, err
	}

	// Generate Header
	if err := g.generateHeader(buf, tmpl); err != nil {
		return nil, err
	}
	// Generate Service
	if err := g.generateServices(buf, tmpl, ps); err != nil {
		return nil, err
	}

	// Generate Messages
	if err := g.generateMessages(buf, tmpl, ps); err != nil {
		return nil, err
	}

	g.genfiles = append(g.genfiles, outfile)
	return g.genfiles, nil
}

func (g *Generator) generateHeader(buf *bufio.Writer, tmpl *template.Template) error {
	data := map[string]interface{}{
		"API":    g.API,
		"Host":   g.Host,
		"Scheme": g.Scheme,
	}
	// write header
	if err := tmpl.ExecuteTemplate(buf, "header", data); err != nil {
		return err
	}
	return buf.Flush()
}

func (g *Generator) generateServices(buf *bufio.Writer, tmpl *template.Template, services []ServiceDefinition) error {
	for _, service := range services {
		data := map[string]interface{}{
			"Name": service.ServiceName(),
			"Rpcs": service.GetRPCs(),
		}
		if err := tmpl.ExecuteTemplate(buf, "service", data); err != nil {
			return err
		}
	}
	return buf.Flush()
}

func (g *Generator) generateMessages(buf *bufio.Writer, tmpl *template.Template, services []ServiceDefinition) error {
	for _, service := range services {
		for _, p := range service.RPCs {
			if len(p.Query) == 0 {
				// use Empty message
				continue
			}
			data := map[string]interface{}{
				"Name":       p.RequestName(),
				"Definition": p.RequestDefinition(),
			}
			if err := tmpl.ExecuteTemplate(buf, "message", data); err != nil {
				return err
			}
			// if Stream is true, defined from other
			if p.Response == nil || p.Response.Stream {
				continue
			}

			// generate Response MediaType
			data = map[string]interface{}{
				"Name":       p.Response.IdentifierName,
				"Definition": p.ResponseDefinition(),
			}
			if err := tmpl.ExecuteTemplate(buf, "message", data); err != nil {
				return err
			}
		}
	}

	return buf.Flush()
}

// Cleanup removes all the files generated by this generator during the last invokation of Generate.
func (g *Generator) Cleanup() {
	for _, f := range g.genfiles {
		os.Remove(f)
	}
	g.genfiles = nil
}

func ensureDelete(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func openFile(path string) (*os.File, error) {
	// clean exist file
	if err := ensureDelete(path); err != nil {
		return nil, err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	return file, nil
}
