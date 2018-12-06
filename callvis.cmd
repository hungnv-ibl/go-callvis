@ECHO OFF
SET PRG=github.com/go-callvis
SET IGNORE=%PRG%/vendor
go-callvis -debug -group pkg,type -nostd -focus main -ignore %IGNORE% github.com\go-callvis