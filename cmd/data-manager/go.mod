module data-manager

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/PlatONEnetwork/PlatONE-Go v0.0.0-fe72c95c689da314dbca1c9a3707f1cc4874ffa6
	github.com/gin-gonic/gin v1.7.1
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	go.mongodb.org/mongo-driver v1.4.1
)

replace github.com/PlatONEnetwork/PlatONE-Go => ../..
