## Инструкция по использованию
Генерация protobuf 

```bash
protoc --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb  --go-grpc_opt=paths=source_relative ./proto/api/*proto
```


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

