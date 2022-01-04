cd manager
go build .
cd ..
cd linter
go build .
cd ..

mkdir -p bin
cp manager/manager bin/manager
cp linter/linter bin/python-linter-1.0
cp linter/linter bin/java-linter-1.0
cp linter/linter bin/python-linter-2.0
cp linter/linter bin/java-linter-2.0
