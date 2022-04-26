define initdir
	@rm -rf ./bin
	@mkdir -vp ./bin/web/public/bootstrap-4.6.0-dist/
	@cp -r ../../web/public/bootstrap-4.6.0-dist/ ./bin/web/public/bootstrap-4.6.0-dist/
	@mkdir -vp ./bin/web/public/views/
	@cp -r ../../web/views/ ./bin/web/views/
	@mkdir -vp ./bin/configs/
	@cp -r ../../web/default/config/ ./bin/configs/
	@mkdir -vp ./bin/cmd/web
endef

main: ../../cmd/web/main.go
	${initdir}
	@go build  -o ./bin/cmd/web/main ../../cmd/web/main.go
	@GOOS=linux GOARCH=amd64 go build  -o ./bin/cmd/web/main_linux64 ../../cmd/web/main.go

.PHONY: image
image :
	${initdir}
	@GOOS=linux GOARCH=amd64 go build  -o ./bin/cmd/web/main_linux64 ../../cmd/web/main.go

