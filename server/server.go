package server

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"ipeye-server/acl"
	"ipeye-server/config"
	"ipeye-server/recorder"
)

const (
	DEFAULT_READ_TIMEOUT  = 90 * time.Second
	DEFAULT_WRITE_TIMEOUT = 90 * time.Second
)

var (
	noDeadline = time.Time{}
)

func Run(config *config.Conf) {
	server, err := net.Listen("tcp", config.Listen.Server)
	if err != nil || server == nil {
		log.Fatalf("[ERROR] Cannot listen: %+v", err)
	}

	log.Println("[INFO] Listening and serving Cloud on " + config.Listen.Server)

	go router()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("[ERROR] Accept failed %+v", err)
			continue
		}
		go handleConnection(conn, config)
	}
}

func handleConnection(conn net.Conn, config *config.Conf) {
	defer conn.Close()

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(DEFAULT_READ_TIMEOUT))
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("[ERROR] Read failed %+v", err)
		return
	}

	conn.SetWriteDeadline(noDeadline)

	if n > 20 {
		switch {
		case strings.Contains(string(buf[:9]), "REGISTER="):
			handleIpc(conn, config, buf[9:n])
		case strings.Contains(strings.ToUpper(string(buf[:20])), "RTSP"):
			handleClient(conn, buf[:n])
		}
	}
}

type (
	register struct {
		CloudID  string `json:"cloudid"`
		Login    string `json:"login"`
		Password string `json:"password"`
		Uri      string `json:"uri"`
		Model    string `json:"model"`
		Vendor   string `json:"vendor"`
		conn     net.Conn
		client   chan net.Conn
	}
)

func handleIpc(conn net.Conn, config *config.Conf, buffer []byte) {
	var register register
	err := json.Unmarshal(buffer, &register)
	if err != nil {
		log.Printf("[ERROR] Unmarshal handleIpc %+v", err)
		return
	}

	permit, need_record := acl.Check(config, register.Login, register.Password, register.CloudID, register.Model, register.Vendor)

	// Нет разрешения на регистрацию
	if !permit {
		log.Printf("[INFO] cloud_id: %s does not have access", register.CloudID)
		return
	}

	path := strings.ReplaceAll(register.CloudID, "/", "-")

	rtsp_url := "rtsp://" + conn.LocalAddr().String() + "/" + path + register.Uri + "?ipc_id=" + path
	log.Printf("[INFO] cloud_id: %s rtsp_url: %s", register.CloudID, rtsp_url)

	register.CloudID = path

	defer func() {
		chanIpcDisconnects <- register.CloudID
	}()

	register.conn = conn
	register.client = make(chan net.Conn)
	chanIpcConnect <- register

	// Нажна запись
	if need_record {
		go recorder.Record(config, register.CloudID, rtsp_url)
	}

	for {
		select {
		case client := <-register.client:
			log.Printf("[INFO] Connect client %s <- ipc %s", client.RemoteAddr().String(), conn.RemoteAddr().String())
			io.Copy(client, conn)
		}
	}
}

func handleClient(conn net.Conn, buffer []byte) {
	r := strings.Split(string(buffer), "\r\n")
	req := strings.Split(strings.TrimSpace(r[0]), " ")
	if len(req) == 3 {
		u, err := url.Parse(req[1])
		if err != nil {
			log.Printf("[ERROR] handleClient parse url %+v", err)
			return
		}

		if u.RawQuery != "" {
			ipc_id := u.Query().Get("ipc_id")
			if ipc_id != "" {
				log.Printf("[INFO] handleClient ipc_id %#v", ipc_id)
				if ipc, ok := ipcs[ipc_id]; ok {
					ipc.conn.SetWriteDeadline(time.Now().Add(DEFAULT_WRITE_TIMEOUT))
					_, err := ipc.conn.Write(buffer)
					if err != nil {
						log.Printf("[ERROR] handleClient write %+v", err)
						return
					}

					ipc.conn.SetWriteDeadline(noDeadline)

					log.Printf("[INFO] Connect ipc %s <- client %s", ipc.conn.RemoteAddr().String(), conn.RemoteAddr().String())
					ipc.client <- conn
					io.Copy(ipc.conn, conn)
				}
			}
		}
	}

	log.Printf("[ERROR] handleClient r %s", r)
}

var (
	// Каналы сообщений
	chanIpcConnect     = make(chan register, 1000)
	chanIpcDisconnects = make(chan string, 1000)

	// Карта с камерами
	ipcs = make(map[string]register)
)

func router() {
	for {
		select {
		case req := <-chanIpcConnect:
			log.Printf("[INFO] Register %#v", req)
			ipcs[req.CloudID] = req
		case id := <-chanIpcDisconnects:
			log.Printf("[INFO] Unregister %#v", id)
			delete(ipcs, id)
		}
	}
}
