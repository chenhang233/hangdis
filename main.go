package main

import "hangdis/tcp"

const banner = `
##################################################
			hangDis	       
##################################################
`

func main() {
	print(banner)
	s, err := tcp.New()
	if err != nil {
		panic(err)
	}
	s.Log.Info.Println("server close...")
}
