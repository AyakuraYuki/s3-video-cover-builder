distFile=s3-video-cover-builder.zip

release:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build main.go \
	&& zip ${distFile} main \
	&& rm main

cleanup:
	rm -f ${distFile}
