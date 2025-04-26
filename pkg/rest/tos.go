package tos

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"

	"golang.org/x/text/language"
)

type TalerTosConfig struct {
  // The default file type
  DefaultFileType string

  // Supported file types
  SupportedFileTypes []string

  // Default language
  DefaultLanguage string
}

func ServiceTermsResponse(w http.ResponseWriter, r *http.Request, termsdatahome string, cfg TalerTosConfig) {
	fileType := cfg.DefaultFileType
	termsLocation := termsdatahome
	for _, typ := range r.Header["Accept"] {
		for _, a := range cfg.SupportedFileTypes {
			if typ == a {
				fileType = a
			}
		}
	}

	if len(r.Header.Get("Accept-Language")) != 0 {
		acceptLangs, _, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
		for _, lang := range acceptLangs {
			extensions, _ := mime.ExtensionsByType(fileType)
			for _, ext := range extensions {
				docFile := fmt.Sprintf("%s/%s/0%s", termsLocation, lang.String(), ext)
				log.Printf("Trying %s\n", docFile)
				fileBytes, err := os.ReadFile(docFile)
				if nil == err {
					w.Header().Set("Content-Type", fileType)
					w.Write(fileBytes)
					return
				}
			}
		}
	}
	// Default document in expected/default format
	defaultLanguage := cfg.DefaultLanguage
	extensions, _ := mime.ExtensionsByType(fileType)
	for _, ext := range extensions {
		docFile := fmt.Sprintf("%s/%s/0%s", termsLocation, defaultLanguage, ext)
		log.Println("Trying " + docFile)
		fileBytes, err := os.ReadFile(docFile)
		if nil == err {
			w.Header().Set("Content-Type", fileType)
			w.Write(fileBytes)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}


func PrivacyPolicyResponse(w http.ResponseWriter, r *http.Request, policydatahome string, cfg TalerTosConfig) {
	fileType := cfg.DefaultFileType
	termsLocation := policydatahome
  for _, typ := range r.Header["Accept"] {
    for _, a := range cfg.SupportedFileTypes {
      if typ == a {
        fileType = a
      }
    }
  }

  if len(r.Header.Get("Accept-Language")) != 0 {
    acceptLangs, _, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
    for _, lang := range acceptLangs {
      extensions, _ := mime.ExtensionsByType(fileType)
      for _, ext := range extensions {
        docFile := fmt.Sprintf("%s/%s/0%s", termsLocation, lang.String(), ext)
        log.Printf("Trying %s\n", docFile)
        fileBytes, err := os.ReadFile(docFile)
        if nil == err {
          w.Header().Set("Content-Type", fileType)
          w.Write(fileBytes)
          return
        }
      }
    }
  }
  // Default document in expected/default format
  defaultLanguage := cfg.DefaultLanguage
  extensions, _ := mime.ExtensionsByType(fileType)
  for _, ext := range extensions {
    docFile := fmt.Sprintf("%s/%s/0%s", termsLocation, defaultLanguage, ext)
    fileBytes, err := os.ReadFile(docFile)
    if nil == err {
      w.Header().Set("Content-Type", fileType)
      w.Write(fileBytes)
      return
    }
  }
  w.WriteHeader(http.StatusNotFound)
}
