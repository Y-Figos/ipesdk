package adapters

import (
	"bytes"
	"fmt"
	"text/template"
	
	"github.com/emersion/go-message/mail"
)
type EMLWriter struct {
	OutputDir	string
	Args		map[string]any
}

func (e *EMLWriter) BuildTmpl(tmpl string, raw map[string]any) (string, error) {
	strData := make(map[string]string, len(raw))
	for k, v := range raw {
		strData[k] = fmt.Sprintf("%v", v)
	}
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, strData); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (e *EMLWriter) BuildEmail() {
	header := mail.Header{}
	header.SetSubject(e.Args["subject"].(string)) 
}

type BatchEmailTemplateWriter struct {
	AttachmentFilePaths  []string
	Template 			 string
	Writer				 *EMLWriter
}

func (e *BatchEmailTemplateWriter) Open() error {
	
	return nil
}	

