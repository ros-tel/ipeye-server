package recorder

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"ipeye-server/config"
)

var (
	// Карта в которой хранятся все процессы записи
	records sync.Map
)

func Record(config *config.Conf, cloud_id string, rtsp_url string) {
	var err error

	// Если уже есть процесс просто выходим
	v, ok := records.Load(cloud_id)
	if ok {
		log.Printf("[INFO] Record process exists for cloud_id: %s url: %s", cloud_id, v)
		return
	}

	// Переходим в корень записей
	err = os.Chdir(config.Recorder.BaseDir)
	if err != nil {
		log.Printf("[ERROR] Record Chdir %+v", err)
		return
	}

	create_record_dir := exec.Command("mkdir", "-p", cloud_id)
	if err := create_record_dir.Run(); err != nil {
		log.Fatalf("[ERROR] Record %+v", err)
	}

	// Переходим в целевую директорию
	err = os.Chdir(config.Recorder.BaseDir + string(os.PathSeparator) + cloud_id)
	if err != nil {
		log.Printf("[ERROR] Record Chdir %+v", err)
		return
	}

	params := strings.Split(config.Recorder.Params, " ")

	// Добавляем в параметры нужные
	params = append(params, "-rtsp_transport")
	params = append(params, "tcp")
	params = append(params, "-i")
	params = append(params, rtsp_url)

	rec := exec.Command(config.Recorder.Cmd, params...)
	err = rec.Start()
	if err != nil {
		log.Printf("[ERROR] Record rec Start %+v", err)
		return
	}

	// log.Printf("[INFO] %v", params)

	// Сохраняем в карту что ведем запись
	records.Store(cloud_id, rtsp_url)

	err = rec.Wait()
	if err != nil {
		log.Printf("[INFO] Command finished with error: %v", err)
	}

	// Удаляем из карты что запись ведется
	records.Delete(cloud_id)

	log.Printf("[INFO] Command finished for cloud_id: %s", cloud_id)
}
