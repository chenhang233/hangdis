package main

import "hangdis/tcp"

const banner = `
##################################################
	hangDis	       
##################################################
`

func main() {
	print(banner)
	s := &tcp.Server{}
	s.New()
}
