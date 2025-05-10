## Как работает сервис
Клиент вызывает метод Subscribe через gRPC, указывая ключ (topic), и получает поток событий (pb.Event), содержащих данные, опубликованные для этого ключа.
Другие клиенты могут публиковать сообщения через метод Publish, указывая ключ и данные. Сообщения асинхронно доставляются всем подписчикам данного ключа.
Pub/Sub-механизм использует каналы Go с буфером (размер задается в конфигурации) для передачи сообщений от публикаторов к подписчикам.
При завершении работы сервис корректно закрывает все подписки, завершает горутины и останавливает gRPC-сервер, используя Graceful Shutdown.

Сервис состоит из следующих основных компонентов:

- **gRPC-сервер (internal/services):**\
Реализует два метода: Publish и Subscribe, определенные в протобуф-описании (pb/proto/api/api.proto).
Метод Publish принимает запросы с ключом и данными, публикуя их в соответствующую тему.
Метод Subscribe создает серверный поток (server streaming), через который клиент получает сообщения для указанного ключа.
Использует библиотеку google.golang.org/grpc для обработки gRPC-запросов.
- **Pub/Sub-механизм (internal/subpub):**\
Реализует асинхронную систему публикации-подписки.
Поддерживает создание подписок на ключи (topics) с асинхронной доставкой сообщений через каналы Go (chan).
Обеспечивает конкурентную обработку подписок и публикаций с использованием мьютексов (sync.Mutex) для безопасного доступа к общим ресурсам.
Поддерживает корректное завершение подписок через метод Unsubscribe и закрытие системы через метод Close.
- **Конфигурация (internal/config):**\
Загружает настройки из YAML-файла и переменных окружения с использованием библиотеки github.com/ilyakaznacheev/cleanenv.
Позволяет задавать параметры, такие как порт gRPC-сервера (GRPC_PORT) и размер буфера подписок (BUFFER_SIZE).
- **Точка входа (cmd/server/main.go):**\
Инициализирует конфигурацию, Pub/Sub-механизм и gRPC-сервер.
Настраивает Graceful Shutdown для корректного завершения работы при получении сигналов ОС (например, SIGINT, SIGTERM).

## Используемые паттерны:

- **Dependency Injection:**\
Компоненты сервиса (например, subpub.SubPub и services.Server) создаются с явной передачей зависимостей через конструкторы (например, NewServer(subpub)). Это упрощает тестирование и замену реализаций.
- **Graceful Shutdown:**\
Сервис обрабатывает сигналы ОС (SIGINT, SIGTERM) и корректно завершает работу, закрывая gRPC-сервер и Pub/Sub-систему с использованием контекста с таймаутом. Это гарантирует, что все активные подписки завершаются, а сообщения обрабатываются до остановки.
- **Clean Architecture:**\
Код организован в пакеты (internal/config, internal/subpub, internal/services), разделяя конфигурацию, бизнес-логику и транспортный слой. Это улучшает читаемость и поддерживаемость.
- **Concurrency Safety:**\
Использование мьютексов (sync.Mutex) и sync.WaitGroup для безопасной конкурентной работы с подписками и публикациями.

## Сборка проекта

Склонируйте проект
```bash
git clone https://github.com/weebatt/asyn-subpub-service.git
cd asyn-subpub-service
```

Обновите зависимости 
```bash
go mod tidy
```

Проверьте что утилита protoc установлена
```bash
protoc --version
```

Сгенерируйте код из subpub.proto 
```bash
protoc --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb  --go-grpc_opt=paths=source_relative ./proto/api/*proto
```

Соберите проект
```bash
go build -o bin/server cmd/server/main.go
```

##  Запуск сервиса
Создайте по примеру config/config.example.yaml config.yaml 

Запустите сервис
```bash
./bin/server -config config.yaml
```

Также предусмотрено развертывание сервиса в Docker
```bash
docker build -t asyn-subpub-service .
docker run -p 50051:50051 asyn-subpub-service
```

### Тестирование

Чтобы запустить тесты 
```bash 
go test ./...
```
Для получения подробной информации по тестированию настоятельно рекомендуется использовать ключи -cover и -v

## Задание состоит из 2 частей.

### Часть 1. Реализация пакета subpub

Требуется реализовать пакет subpub - простую шину событий, работающую по принципу Publisher-Subscriber.

### Требования к шине:
- На один subject может подписываться (и отписываться) множество подписчиков
- Один медленный подписчик не должен тормозить остальных
- Нельзя терять порядок сообщений (FIFO очередь)
- Метод Close должен учитывать переданный контекст. Если он отменен - выходим сразу, работающие хендлеры оставляем работать
- Горутины (если они будут) течь не должны

### API пакета subpub:

```go
package subpub

import "context"

// MessageHandler is a callback function that process massages delivered to subscribers.
type MessageHandler func(msg interface{})

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription is for. 
	Unsubscribe()
}

type SubPub interface {
	// Subscribe creates an asynchronous queue subscribers on the given subject. 
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	// Publish publishes the msg argument to the give subject.  
	Publish(subject string, msg interface{}) error

	// Close will shutdown the sub-pub system.
	// May be blocked by data deliver until the context is canceled. 
	Close(ctx context.Context) error
}

func NewSubPub() SubPub {
	panic("Implement me")
}
```

### Часть 2. Реализация сервиса подписок
Во второй части задания требуется с использованием 
пакета subpub из 1 части реализовать сервис подписок. 
Сервис работает по gRPC. Есть возможность подписаться 
на события по ключу и опубликовать события по ключу для всех подписчиков.
         
### Protobuf-схема gRPC сервиса:
```go
syntax = "proto3";

import "google/protobuf/empty.proto";

service PubSub {
rpc Subscribe(SubscribeRequest) returns (stream Event);

rpc Publish(PublishRequest) returns (google.protobuf.Empty);
}

message SubscribeRequest {
string key = 1;
}

message PublishRequest {
string key = 1;
string data = 2;
}

message Event {
string data = 1;
}
```
### Дополнительные требования
Также пользуйся стандартными статус-кодами gRPC из пакетов
```aiignore
google.golang.org/grpc/status
```
и
```aiignore
google.golang.org/grpc/codes
```

### Ожидаемое в решении
В качестве критериев успешности и неуспешности запросов к сервису.\
Что еще ожидается в решении:
- Обязательно должно быть описание того, как работает сервис и как его собирать.
- У сервиса должен быть свой конфиг, куда можно прописать порты и прочие параметры (на ваше усмотрение).
- Логирование.
- Приветствуется использование известных паттернов при разработке микросервисов на Go (например, dependency injection, graceful shutdown и пр.). Если таковые будут
  использоваться, то просьба упомянуть его в описании решения.

