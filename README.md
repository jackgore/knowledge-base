# go-service

go-service is a base repository to quickly develop micro-services using Go. 
It provides functionality and structure common to Go micro-services. 

## Getting Started

1) Clone the repository to a new folder within your go path

```
$ pwd
/Users/jack/go/src/github.com/JonathonGore
$ git clone --depth 1 git@github.com:JonathonGore/go-service.git REPOSITORY && cd REPOSITORY
```

2) Update git to point to the GitHub repository you want to use (if using ssh use corresponding URL instead)
```
git remote set-url origin https://github.com/USERNAME/REPOSITORY.git
```

3) Update import paths
```
./scripts/setup.sh 'github.com/USERNAME/REPOSITORY'
```
4) Get dependencies
```
dep ensure
```
5a) Run the project
```
 go run main.go
```
5b) Or run with Docker
```
docker build -t go-service . && docker run -p 3000:3000 go-service
```

Now you should be able to hit the test endpoint at `http://0.0.0.0:3000/v1/go-service`.
