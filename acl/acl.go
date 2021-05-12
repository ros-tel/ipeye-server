package acl

import "ipeye-server/config"

// Проверяет имеет ли камера доступ к регистрации и нужна ли запись
func Check(config *config.Conf, login, password, cloud_id, model, vendor string) (bool, bool) {
	// Временно возвращаем что есть доступ и нужна запись
	return true, true
}
