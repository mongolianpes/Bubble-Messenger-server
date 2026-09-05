package writer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"server/internal/gzip"
	"sync"
	"time"
)

type FileWorker struct {
	ch       chan []byte
	stop     chan struct{}
	lastUsed time.Time
}

type FileWriterManager struct {
	mu      sync.Mutex
	workers map[string]*FileWorker
	timeout time.Duration
}

func NewFileWriterManager(timeout time.Duration) *FileWriterManager {
	return &FileWriterManager{
		workers: make(map[string]*FileWorker),
		timeout: timeout,
	}
}

// ---------------------------
// ПУБЛИЧНАЯ ФУНКЦИЯ ЗАПИСИ
// ---------------------------

func (m *FileWriterManager) Write(pathDir string, data []byte) {
	path := filepath.Join(pathDir, "newMessages")

	m.mu.Lock()
	worker, exists := m.workers[path]
	if !exists {
		worker = &FileWorker{
			ch:       make(chan []byte, 1000),
			stop:     make(chan struct{}),
			lastUsed: time.Now(),
		}
		m.workers[path] = worker
		go m.startWorker(path, worker)
	}
	worker.lastUsed = time.Now()
	m.mu.Unlock()

	worker.ch <- data
}

// ---------------------------
// ПУБЛИЧНАЯ ФУНКЦИЯ ЧТЕНИЯ
// ---------------------------

func (m *FileWriterManager) Read(pathDir string) ([]byte, error) {
	path := filepath.Join(pathDir, "newMessages")

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return []byte{}, nil
		}
		return nil, err
	}

	return os.ReadFile(path)
}

// ---------------------------
// ВОРКЕР ЗАПИСИ
// ---------------------------

func (m *FileWriterManager) startWorker(path string, w *FileWorker) {
	os.MkdirAll(filepath.Dir(path), 0700)

	// ВАЖНО: O_RDWR — можно читать и писать, Seek работает корректно
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-w.ch:
			w.lastUsed = time.Now()

			// Читаем старые данные
			file.Seek(0, 0)
			oldMessages, err := io.ReadAll(file)
			if err != nil {
				oldMessages = []byte{}
			}

			// Формируем новое сообщение
			newMessage := append([]byte("\n"), msg...)

			// Если новое сообщение само по себе слишком большое
			if len(newMessage) > 400 {
				m.rotateSingle(path, newMessage)
				continue
			}

			// Комбинированные данные
			combined := append(append([]byte{}, oldMessages...), newMessage...)

			// Если combined < 800 — просто записываем
			if len(combined) < 800 {
				file.Truncate(0)
				file.Seek(0, 0)
				file.Write(combined)
				continue
			}

			// combined >= 800 → проверяем gzip(combined)
			gzCombined, err := gzip.CompressGzip(combined)
			if err != nil {
				return
			}

			if len(gzCombined) <= 800 {
				m.saveArchive(path, gzCombined, true)
				file.Truncate(0)
				file.Seek(0, 0)
				continue
			}

			// gzip(combined) > 800 → архивируем только старые данные
			if len(oldMessages) > 0 {
				gzOld, err := gzip.CompressGzip(oldMessages)
				if err != nil {
					return
				}
				m.saveArchive(path, gzOld, true)
			}

			// В файл пишем только новое сообщение
			file.Truncate(0)
			file.Seek(0, 0)
			file.Write(msg)

		case <-ticker.C:
			if time.Since(w.lastUsed) > m.timeout {
				m.mu.Lock()
				delete(m.workers, path)
				m.mu.Unlock()
				return
			}
		}
	}
}

// ---------------------------
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ---------------------------

// Архивация только нового сообщения
func (m *FileWriterManager) rotateSingle(path string, newMsg []byte) {
	dir := filepath.Dir(path)
	timeNow := time.Now()

	fileName := fmt.Sprintf(
		"%d.%d.%d.%d.%d.%d.%d",
		timeNow.Year(),
		int(timeNow.Month()),
		timeNow.Day(),
		timeNow.Hour(),
		timeNow.Minute(),
		timeNow.Second(),
		timeNow.Nanosecond(),
	)

	newPath := filepath.Join(dir, fileName)

	gzNew, err := gzip.CompressGzip(newMsg)
	if err != nil {
		return
	}

	if len(gzNew) < len(newMsg) {
		newPath += ".gz"
		os.WriteFile(newPath, gzNew, 0644)
	} else {
		os.WriteFile(newPath, newMsg, 0644)
	}
}

// Сохранение архива
func (m *FileWriterManager) saveArchive(path string, data []byte, gz bool) {
	dir := filepath.Dir(path)
	timeNow := time.Now()

	fileName := fmt.Sprintf(
		"%d.%d.%d.%d.%d.%d.%d",
		timeNow.Year(),
		int(timeNow.Month()),
		timeNow.Day(),
		timeNow.Hour(),
		timeNow.Minute(),
		timeNow.Second(),
		timeNow.Nanosecond(),
	)

	if gz {
		fileName += ".gz"
	}

	newPath := filepath.Join(dir, fileName)
	os.WriteFile(newPath, data, 0644)
}

// package writer

// import (
// 	"fmt"
// 	"io"
// 	"os"
// 	"path/filepath"
// 	"server/gzip"
// 	"sync"
// 	"time"
// )

// type FileWorker struct {
// 	ch       chan []byte
// 	stop     chan struct{}
// 	lastUsed time.Time
// }

// type FileWriterManager struct {
// 	mu      sync.Mutex
// 	workers map[string]*FileWorker
// 	timeout time.Duration
// }

// func NewFileWriterManager(timeout time.Duration) *FileWriterManager {
// 	return &FileWriterManager{
// 		workers: make(map[string]*FileWorker),
// 		timeout: timeout,
// 	}
// }

// // pathDir — путь к директории, например: "users/john/messages/"
// func (m *FileWriterManager) Write(pathDir string, data []byte) {
// 	// Полный путь к файлу newMessages
// 	path := filepath.Join(pathDir, "newMessages")

// 	m.mu.Lock()
// 	worker, exists := m.workers[path]
// 	if !exists {
// 		worker = &FileWorker{
// 			ch:       make(chan []byte, 1000),
// 			stop:     make(chan struct{}),
// 			lastUsed: time.Now(),
// 		}
// 		m.workers[path] = worker
// 		go m.startWorker(path, worker)
// 	}
// 	worker.lastUsed = time.Now()
// 	m.mu.Unlock()

// 	worker.ch <- data
// }

// func (m *FileWriterManager) startWorker(path string, w *FileWorker) {
// 	// Создаём директорию, если её нет
// 	os.MkdirAll(filepath.Dir(path), 0755)

// 	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
// 	if err != nil {
// 		return
// 	}
// 	defer file.Close()

// 	ticker := time.NewTicker(5 * time.Second)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case msg := <-w.ch:
// 			oldMessages, err := io.ReadAll(file)
// 			if err != nil {
// 				return
// 			}

// 			newMessage := append([]byte("\n"), msg...)

// 			if len(newMessage) > 400 {
// 				dir := filepath.Dir(path)
// 				timeNow := time.Now()
// 				fileName := fmt.Sprintf(
// 					"%d.%d.%d.%d.%d.%d.%d",
// 					timeNow.Year(),
// 					int(timeNow.Month()),
// 					timeNow.Day(),
// 					timeNow.Hour(),
// 					timeNow.Minute(),
// 					timeNow.Second(),
// 					timeNow.Nanosecond(),
// 				)
// 				newName := filepath.Join(dir, fileName)

// 				gzipNewMessages, err := gzip.CompressGzip(newMessage)
// 				if err != nil {
// 					return
// 				}

// 				if len(gzipNewMessages) < len(oldMessages) {
// 					newName += ".gz"
// 					if err := os.WriteFile(newName, gzipNewMessages, 0644); err != nil {
// 						return
// 					}
// 					return
// 				}

// 				os.WriteFile(newName, newMessage, 0644)
// 				return
// 			}

// 			combinedMessages := append(append([]byte{}, oldMessages...), append([]byte{}, newMessage...)...)

// 			if len(combinedMessages) > 800 {
// 				gzipCombinedMessages, err := gzip.CompressGzip(combinedMessages)
// 				if err != nil {
// 					return
// 				}
// 				if len(gzipCombinedMessages) > 800 {
// 					gzipOldMessages, err := gzip.CompressGzip(oldMessages)
// 					if err != nil {
// 						return
// 					}

// 					dir := filepath.Dir(path)
// 					timeNow := time.Now()
// 					fileName := fmt.Sprintf(
// 						"%d.%d.%d.%d.%d.%d.%d",
// 						timeNow.Year(),
// 						int(timeNow.Month()),
// 						timeNow.Day(),
// 						timeNow.Hour(),
// 						timeNow.Minute(),
// 						timeNow.Second(),
// 						timeNow.Nanosecond(),
// 					)
// 					newName := filepath.Join(dir, fileName)

// 					if len(gzipOldMessages) < len(oldMessages) {
// 						newName += ".gz"
// 						if err := os.WriteFile(newName, gzipOldMessages, 0644); err != nil {
// 							return
// 						}
// 						return
// 					}

// 					if err = os.WriteFile(newName, oldMessages, 0644); err != nil {
// 						return
// 					}

// 					file.Truncate(0)
// 					file.Seek(0, 0)

// 					//тут тотальная проверка нового сообщения (примерно тоже самое что и выше с комбинированным было)

// 					file.Write(newMessage)
// 				} else {
// 					file.Write(newMessage)
// 				}

// 			} else {
// 				file.Write(newMessage)
// 			}

// 		case <-ticker.C:
// 			if time.Since(w.lastUsed) > m.timeout {
// 				m.mu.Lock()
// 				delete(m.workers, path)
// 				m.mu.Unlock()
// 				return
// 			}
// 		}
// 	}
// }

// how to use
// manager := writer.NewFileWriterManager(60 * time.Second)

// // из любой горутины
// manager.Write("/var/log/a.txt", "hello")
// manager.Write("/var/log/b.txt", "world")

// var gzipWriterPool = sync.Pool{
// 	New: func() any {
// 		return gzip.NewWriter(nil)
// 	},
// }
// var bufferPool = sync.Pool{
// 	New: func() any {
// 		return new(bytes.Buffer)
// 	},
// }

// func CompressGzip(data []byte) ([]byte, error) {
// 	buf := bufferPool.Get().(*bytes.Buffer)
// 	buf.Reset()

// 	gz := gzipWriterPool.Get().(*gzip.Writer)
// 	gz.Reset(buf)

// 	_, err := gz.Write(data)
// 	if err != nil {
// 		gzipWriterPool.Put(gz)
// 		bufferPool.Put(buf)
// 		return nil, err
// 	}

// 	if err := gz.Close(); err != nil {
// 		gzipWriterPool.Put(gz)
// 		bufferPool.Put(buf)
// 		return nil, err
// 	}

// 	out := make([]byte, buf.Len())
// 	copy(out, buf.Bytes())

// 	gzipWriterPool.Put(gz)
// 	bufferPool.Put(buf)

// 	return out, nil
// }
