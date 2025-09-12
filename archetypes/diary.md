+++
date = '{{ .Date }}'
draft = false
title = '{{ replace .File.ContentBaseName "-" " " | title }}'
authors = [""]
topics = [""]
opening = ""
image = "{{- replace .File.ContentBaseName ":" "" -}}."
+++
