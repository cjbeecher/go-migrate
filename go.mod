module go-migrate

go 1.18

require gomigrate v0.0.0

require github.com/lib/pq v1.10.9 // indirect

replace gomigrate v0.0.0 => ./src/gomigrate
