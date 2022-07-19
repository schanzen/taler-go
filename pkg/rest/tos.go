package tos

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"strings"

	"golang.org/x/text/language"
	"gopkg.in/ini.v1"
)

func ServiceTermsResponse(s *ini.Section, w http.ResponseWriter, r *http.Request) {
	fileType := s.Key("default_doc_filetype").MustString("text/html")
	termsLocation := s.Key("default_tos_path").MustString("terms/")
	for _, typ := range r.Header["Accept"] {
		for _, a := range strings.Split(s.Key("supported_doc_filetypes").String(), " ") {
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
				fileBytes, err := ioutil.ReadFile(docFile)
				if nil == err {
					w.Header().Set("Content-Type", fileType)
					w.Write(fileBytes)
					return
				}
			}
		}
	}
	// Default document in expected/default format
	defaultLanguage := s.Key("default_doc_lang").MustString("en")
	extensions, _ := mime.ExtensionsByType(fileType)
	for _, ext := range extensions {
		docFile := fmt.Sprintf("%s/%s/0%s", termsLocation, defaultLanguage, ext)
		log.Println("Trying " + docFile)
		fileBytes, err := ioutil.ReadFile(docFile)
		if nil == err {
			w.Header().Set("Content-Type", fileType)
			w.Write(fileBytes)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func ServicePrivacyPolicyResponse(s *ini.Section, w http.ResponseWriter, r *http.Request) {
	fileType := s.Key("default_doc_filetype").MustString("text/html")
	termsLocation := s.Key("default_pp_path").MustString("privacy/")
	for _, typ := range r.Header["Accept"] {
		for _, a := range strings.Split(s.Key("supported_doc_filetypes").String(), " ") {
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
				fileBytes, err := ioutil.ReadFile(docFile)
				if nil == err {
					w.Header().Set("Content-Type", fileType)
					w.Write(fileBytes)
					return
				}
			}
		}
	}
	// Default document in expected/default format
	defaultLanguage := s.Key("default_doc_lang").MustString("en")
	extensions, _ := mime.ExtensionsByType(fileType)
	for _, ext := range extensions {
		docFile := fmt.Sprintf("%s/%s/0%s", termsLocation, defaultLanguage, ext)
		fileBytes, err := ioutil.ReadFile(docFile)
		if nil == err {
			w.Header().Set("Content-Type", fileType)
			w.Write(fileBytes)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}
