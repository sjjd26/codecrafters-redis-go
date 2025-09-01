My version of the codecrafters Redis project, done in Go.

The solution mirrors Redis' single command execution thread with an event loop.

Stages completed:
- General string storing
- Replication
- RDB Persistence
- Lists

#### Run instructions
To run just use `go run .` from the app directory.
Optional named arguments: `--port <port_number> --dir <rdb_persistence_directory> --dbfilename <rdb_persistence_filename> --replicaOf <master_host> <master_port>`
