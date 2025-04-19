
## kafka相关指令

开启kafka
zookeeper-server-start.bat ..\..\config\zookeeper.properties

kafka-server-start.bat ..\..\config\server.properties


```shell

windows环境下：
// 创建一个主题
D:\devtools\kafka_2.13-3.4.0\bin\windows>kafka-topics.bat --bootstrap-server 127.0.0.1:9092 --topic topic-like-count --create
./kafka-topics.sh --create --topic topic-like-count --bootstrap-server localhost:9092    // linux
    
// 查看所有主题 
D:\devtools\kafka_2.13-3.4.0\bin\windows>kafka-topics.bat --bootstrap-server 127.0.0.1:9092 --list
./kafka-topics.sh --bootstrap-server localhost:9092 --list    // linux
./kafka-topics.sh --describe --topic topic-beyond-like --bootstrap-server localhost:9092  // linux 查看某个主题信息

// 启动生产者
D:\devtools\kafka_2.13-3.4.0\bin\windows>kafka-console-producer.bat --broker-list localhost:9092 --topic topic-like-count
./kafka-console-producer.sh --topic topic-like-count --bootstrap-server localhost:9092

//启动消费者
D:\devtools\kafka_2.13-3.4.0\bin\windows> kafka-console-consumer.bat --bootstrap-server localhost:9092 --topic topic-like-count --from-beginning
./kafka-console-consumer.sh --topic topic-beyond-like --from-beginning --bootstrap-server localhost:9092

```