# Plano experimental inicial — TCC MQTT

## Arquitetura

- Gerador de carga: computador local do pesquisador.
- Broker EMQX: EC2 `tcc-emqx`.
- Broker Mosquitto: EC2 `tcc-mosquitto`.
- Ambas as EC2 derivadas da mesma AMI e com a mesma configuração de hardware e sistema operacional.
- O backend, PostgreSQL e frontend ficam fora da fase de benchmark dos brokers.

## Categorias de métricas

### 1. Desempenho

- Latência fim a fim entre publicação e recebimento no gerador local.
- Média, mínimo, máximo, p50, p95 e p99 da latência.
- Throughput efetivo de publicação e recebimento (mensagens/s).
- Taxa de entrega.
- Mensagens perdidas.
- Duplicatas observadas.

### 2. Consumo de recursos

Coletado diretamente na EC2 do broker:

- CPU da instância (`sar`).
- Memória da instância (`sar`).
- Rede (`sar -n DEV`).
- CPU e memória/RSS do processo do broker (`pidstat`).
- Número de conexões TCP estabelecidas na porta 1883.

### 3. Volume de conexões

- Número solicitado de conexões simultâneas.
- Conexões estabelecidas.
- Falhas de conexão.
- Taxa efetiva de estabelecimento de conexões.
- Latência de conexão p50/p95/p99.
- CPU e RAM do broker durante o teste.

## Condições padronizadas

- MQTT 5.0.
- Mesma máquina geradora de carga.
- Mesma conexão de Internet.
- Mesma EC2 para cada função, com hardware equivalente entre brokers.
- QoS de publicação e assinatura iguais em cada cenário.
- Retain desabilitado.
- Sessão limpa e expiração igual a zero nos clientes de benchmark.
- Autenticação desabilitada na comparação inicial.
- Payload padrão inicial: 256 bytes.
- Warm-up antes da janela de medição.
- Dashboard do EMQX fechado durante as execuções oficiais.
- Repetir cada cenário final pelo menos 3 vezes.

## Fase piloto recomendada

Antes da matriz final, executar em cada broker:

| Clientes publicadores | QoS | Taxa por cliente | Payload | Warm-up | Medição |
|---:|---:|---:|---:|---:|---:|
| 100 | 0 | 1 msg/s | 256 B | 10 s | 60 s |
| 100 | 1 | 1 msg/s | 256 B | 10 s | 60 s |
| 100 | 2 | 1 msg/s | 256 B | 10 s | 60 s |

A fase piloto serve para validar o simulador, a coleta de recursos e o procedimento. A matriz final será definida depois de observar onde começam a aparecer diferenças relevantes e limites do gerador/broker.

## Observação sobre latência

Como publisher e subscriber são executados no mesmo computador local e o broker fica na AWS, a métrica é **latência MQTT fim a fim**, incluindo o caminho de rede até a nuvem e de volta. Não deve ser descrita como tempo interno de processamento do broker.
