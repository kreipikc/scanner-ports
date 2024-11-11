## What kind of project is this?
This is a simple port scanner on golang.

## What technologies have I used?
- Golang

All the code is written in golang

## Why did I even start creating this project?
I created this project to study the basic work with the net and the _net_ library, as well as to study **streams** and **goroutines**.

## How usage?
To run the program, just write run the `main.go` file - `go run main.go`

You can also configure settings such as:
- `address` - specify the address you need, _default: **scanme.nmap.org**_ (accepts **string**),
- `first_port` - specify which port the scan will start from, _default: **1**_ (accepts **int**),
- `last_port` - specify the port to which the scan will take place, _default: **1024**_ (accpet **int**),
- `max_buffer` - here, specify the number for the goroutine buffer, <u>change it if you understand why</u>, _default: **100**_ (accept **int**).
- `save_format` - if you want to save the result, then add it in what format, _default: **none**_ (accepts **json**, **txt**)
- `document` - if you need to scan several addresses from a txt file (addresses should be in the column line by line), then specify the path to it

### **!!!Use only address or only document!!!**

Launch example for 1 address:
```bash
go run main.go --address scanme.nmap.org --first_port 1 --last_port 1024 --max_buffer 100 --save_format json
```

Launch example for document with address:
```bash
go run main.go --document file.txt --first_port 1 --last_port 1024 --max_buffer 100 --save_format json
```