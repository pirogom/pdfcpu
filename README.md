# pirogom/pdfcpu

clone from [pdfcpu](https://github.com/pdfcpu)

모두의 PDF 개발에 사용된 PDF 프로세서 입니다.
- 장점 : 빠름
- 단점 :
PDF 스펙을 제대로 지킨 1.7 버전의 파일이 아닌 경우 매우 높은 확률로
PDF파일을 읽는데 실패합니다. 오류나 PDF 1.7 스펙에 맞지 않는 Annotation이 
존재하는 PDF파일들은 다른 무언가로 오류 보정을 한 뒤에 사용해야 합니다.
- 피로곰이 한짓 :
  PDF 메타데이터 수정시 UTF-8 사용이 안되고 Chart-set에 따른 문제가
  발생하는 부분 수정 .. PIROGOM 주석 참고


# pdfcpu: a Go PDF processor
pdfcpu is a PDF processing library written in [Go](http://golang.org) supporting encryption.
It provides both an API and a CLI. Supported are all versions up to PDF 1.7 (ISO-32000).

## Documentation

* The main entry point is [pdfcpu.io](https://pdfcpu.io).
* For CLI examples also go to [pdfcpu.io](https://pdfcpu.io). There you will find explanations of all the commands and their parameters.
* For API examples of all pdfcpu operations please refer to [GoDoc](https://pkg.go.dev/github.com/pirogom/pdfcpu/pkg/api).
* 
## Reminder

* Always make sure your work is based on the latest commit!<br>
* pdfcpu is still *Alpha* - bugfixes are committed on the fly and will be mentioned in the next release notes.<br>
* Follow [pdfcpu](https://twitter.com/pdfcpu) for news and release announcements.
* For quick questions or discussions get in touch on the [Gopher Slack](https://invite.slack.golangbridge.org/) in the #pdfcpu channel.
* 

### Using Go Modules

```
git clone https://github.com/pirogom/pdfcpu
cd pdfcpu/cmd/pdfcpu
go install
pdfcpu version
```

### Run in a Docker container

```
docker build -t pdfcpu .
# mount current folder into container to process local files
docker run -it --mount type=bind,source="$(pwd)",target=/app pdfcpu ./pdfcpu validate -mode strict /app/pdfs/a.pdf
```

## License

Apache-2.0
