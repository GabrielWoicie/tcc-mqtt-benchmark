# TCC MQTT Benchmark

Simulador de carga em Go para comparação acadêmica entre EMQX e Eclipse Mosquitto.

O projeto usa **MQTT 5.0** por meio do cliente oficial Eclipse Paho para Go (`github.com/eclipse/paho.golang`) e possui dois modos independentes:

1. `message`: desempenho, entrega e latência fim a fim.
2. `connections`: volume e velocidade de estabelecimento de conexões simultâneas.

O consumo de CPU/RAM/rede do broker é coletado separadamente na própria EC2 com `scripts/collect_broker_metrics.sh`, evitando misturar métricas do gerador com métricas do servidor.

## Requisitos no Windows

- Go 1.21 ou superior.
- Porta 1883 dos brokers liberada no Security Group para o IP público do computador de teste.

Verifique:

```powershell
go version
```

## Preparação

No diretório do projeto:

```powershell
go mod tidy
go build -o mqttbench.exe .\cmd\mqttbench
```

Teste a ajuda:

```powershell
.\mqttbench.exe help
```

## Configuração por arquivo (recomendado)

Edite [`mqttbench.conf`](mqttbench.conf), que explica cada parâmetro em comentários.
Substitua `IP_PUBLICO_EMQX` pelo IP do broker e ajuste os valores. Depois, na pasta
do projeto, execute:

```powershell
.\mqttbench.exe
```

O arquivo é lido a cada execução: não é preciso recompilar ao mudar parâmetros.
Para compilar o código atualizado uma vez, use
`go build -o mqttbench.exe .\cmd\mqttbench`. Também é possível executar diretamente
com `go run ./cmd/mqttbench`.

O formato é `campo=valor`, sem aspas, com comentários `#` em linhas próprias.
Salve em UTF-8. JSON não aceita comentários nativamente; este formato simples
permite documentar os campos sem adicionar dependências. Um `config.go` exigiria
recompilação a cada alteração.

O arquivo padrão é `mqttbench.conf` na **pasta atual do terminal**. Para escolher
outro arquivo e, opcionalmente, sobrescrever valores nessa execução:

```powershell
.\mqttbench.exe --config .\mqttbench.conf --qos 1 --clients 10
```

Precedência: parâmetros do comando > valores do arquivo > padrões do programa.
`--config` deve ser o primeiro argumento. O campo `mode` escolhe `message` ou
`connections`; para conexões, comente os campos exclusivos de mensagens e
descomente `hold`, conforme indicado no arquivo. Campos desconhecidos, repetidos
ou com tipos inválidos geram erro. Caminhos como `out` são relativos à pasta atual
do terminal. Valores podem conter `=` e `#`; espaços nas extremidades são removidos.

Os comandos abaixo continuam funcionando com seus próprios parâmetros, sem ler
automaticamente o arquivo quando um modo é informado explicitamente.

## 1. Benchmark de mensagens

Exemplo inicial EMQX:

```powershell
.\mqttbench.exe message `
  --broker mqtt://IP_PUBLICO_EMQX:1883 `
  --broker-name emqx `
  --clients 100 `
  --qos 0 `
  --rate 1 `
  --warmup 10s `
  --duration 60s `
  --drain 5s `
  --payload-bytes 256
```

Repita exatamente com o Mosquitto:

```powershell
.\mqttbench.exe message `
  --broker mqtt://IP_PUBLICO_MOSQUITTO:1883 `
  --broker-name mosquitto `
  --clients 100 `
  --qos 0 `
  --rate 1 `
  --warmup 10s `
  --duration 60s `
  --drain 5s `
  --payload-bytes 256
```

Depois repita com `--qos 1` e `--qos 2`.

### Como funciona o modo `message`

- Um cliente MQTT dedicado (`collector`) assina `tcc/benchmark/<run-id>/#`.
- São criadas N conexões de publishers.
- Cada publisher possui Client ID único e publica em seu próprio tópico.
- O payload contém `run_id`, índice do cliente, sequência e `sent_at_ns`.
- O collector usa o relógio da **mesma máquina geradora** para calcular `received_at - sent_at`, evitando necessidade de sincronizar relógios entre PC e EC2.
- O warm-up é enviado normalmente, mas ignorado nas estatísticas.
- Ao terminar a janela de medição há um período de drain para mensagens em trânsito.

Arquitetura do teste:

```text
N publishers (PC local) -----> broker AWS -----> 1 collector (PC local)
```

A latência obtida é fim a fim e inclui rede externa + broker.

## 2. Benchmark de conexões

Exemplo:

```powershell
.\mqttbench.exe connections `
  --broker mqtt://IP_PUBLICO_EMQX:1883 `
  --broker-name emqx `
  --clients 1000 `
  --connect-rate 100 `
  --hold 60s
```

O teste mede:

- conexões solicitadas;
- conexões estabelecidas;
- falhas;
- conexões estabelecidas por segundo;
- latência média, p50, p95, p99, mínimo e máximo para MQTT CONNECT/CONNACK.

As conexões bem-sucedidas ficam abertas durante `--hold` para permitir medir CPU/RAM do broker sob o volume de conexões.

## 3. Coleta de CPU/RAM/rede na EC2

Copie `scripts/collect_broker_metrics.sh` para cada EC2 ou crie o arquivo manualmente.

EMQX:

```bash
chmod +x collect_broker_metrics.sh
./collect_broker_metrics.sh emqx 90 1 ~/tcc-metrics/emqx-piloto-qos0
```

Mosquitto:

```bash
chmod +x collect_broker_metrics.sh
./collect_broker_metrics.sh mosquitto 90 1 ~/tcc-metrics/mosquitto-piloto-qos0
```

Para um teste de `10 s warm-up + 60 s medição + 5 s drain`, 90 segundos dão margem para iniciar o coletor antes do benchmark.

O script gera:

- `sar_cpu.txt`
- `sar_memory.txt`
- `sar_network.txt`
- `pidstat_broker.txt`
- `connections.csv`
- `metadata.txt`

Depois copie para o computador:

```powershell
scp -i "C:\caminho\tcc-iot-key.pem" -r ubuntu@IP_DA_EC2:~/tcc-metrics .\broker-metrics\
```

## 4. Resultados do simulador

Cada execução cria uma pasta independente em `results/`.

Modo `message`:

- `summary.json`: todas as métricas e metadados.
- `summary.csv`: linha resumida pronta para planilha.
- `messages.csv`: uma linha por mensagem recebida, útil para análise estatística e gráficos.

Modo `connections`:

- `summary.json`
- `connections.csv`: sucesso/falha e latência por tentativa.

## 5. Piloto automatizado

O PowerShell `scripts/run_pilot.ps1` executa QoS 0, 1 e 2 para 100 clientes, 1 msg/s por cliente, 60 segundos cada.

Exemplo:

```powershell
.\scripts\run_pilot.ps1 -Broker "mqtt://IP_PUBLICO_EMQX:1883" -BrokerName emqx
```

Execute o mesmo depois contra o Mosquitto.

## Regras para experimentos oficiais

- Usar cabo Ethernet quando possível.
- Fechar downloads, streaming e aplicações que gerem tráfego pesado.
- Não deixar o Dashboard EMQX aberto durante os benchmarks oficiais.
- Rodar EMQX e Mosquitto sob as mesmas condições.
- Não alterar versão/configuração dos brokers entre repetições.
- Registrar ping/qualidade da rede no início de cada sessão de testes.
- Monitorar o computador gerador para verificar se ele não se torna o gargalo.
- Fazer primeiro a fase piloto; não começar diretamente com milhares de conexões.
- Intercalar a ordem dos brokers nas repetições finais para reduzir viés temporal de rede (ex.: EMQX → Mosquitto → Mosquitto → EMQX).

## Próximo passo

Execute primeiro o piloto de 100 clientes. Só depois de validar os resultados definimos a matriz final (100/250/500/1000 clientes, taxas e duração), evitando criar dezenas de testes sem necessidade científica.

## Versão 0.2.0

Ajustes após o primeiro teste piloto real:

- corrigido o limite final da janela de medição para que uma execução de 15 s a 1 msg/s por cliente não contabilize um tick adicional no instante exato de encerramento;
- adicionado `tool_version` aos arquivos `summary.json` dos modos `message` e `connections` para facilitar a rastreabilidade dos experimentos.

Para os benchmarks oficiais, use a mesma versão do executável em todas as execuções de EMQX e Mosquitto.

## Versão 0.3.0: experimentos reproduzíveis

Três formas de executar usando configuração:

```powershell
.\mqttbench.exe
.\mqttbench.exe --config configs\pilot\emqx-qos0.conf
.\mqttbench.exe --config configs\pilot\emqx-qos0.conf --clients=10
```

Sem argumentos, lê `mqttbench.conf` na pasta atual. A prioridade continua sendo
CLI > arquivo > defaults. Os modos explícitos da CLI continuam disponíveis.

`scenario-id` identifica o experimento sem modificar o `run_id`. É opcional;
`repetition` tem default 1 e deve ser >= 1. Ambos aparecem nos resumos JSON e CSV
dos dois modos. Quando há identificador, a pasta usa
`<scenario-id>_<broker>_<timestamp>`: caracteres fora de A–Z, a–z, 0–9, hífen
e sublinhado são substituídos por `_`, sem alterar o identificador no resumo.
Sem identificador, permanece o formato anterior.

O modo `message` usa `publish-pattern=staggered` por padrão. Para cinco
publishers a 1 mensagem/s por publisher:

```text
                       primeiros envios (ms)
staggered:     cliente 0: 0, 1000, 2000
               cliente 1: 200, 1200, 2200
               cliente 2: 400, 1400, 2400
               cliente 3: 600, 1600, 2600
               cliente 4: 800, 1800, 2800
synchronized:  todos:     0, 1000, 2000
```

Os horários são aproximados e dependem do scheduler. O período continua sendo
1/rate por publisher; a 5 msg/s os offsets são 0, 40, 80, 120 e 160 ms.
O relógio e os offsets começam no warmup e seguem sem reinício na medição.
Mensagens de warmup são excluídas das métricas. A janela é `[start, end)`;
5 clientes × 1 msg/s × 15 s resultam em 75 tentativas quando o gerador acompanha
a taxa. Slots perdidos por atraso não geram rajadas de compensação.
O modo `connections` rejeita `publish-pattern` se informado explicitamente.

Cada pasta de execução recebe `effective-config.conf`, com valores resolvidos,
incluindo defaults e overrides. Senhas configuradas aparecem como
`password=<redacted>`; os resumos também não expõem a senha.
O arquivo é um registro da execução: `tool-version` é metadado e a senha está
ocultada. Para reutilizá-lo com `--config`, remova `tool-version` e restaure
a credencial, se necessária.

Os resumos de mensagens preservam as colunas anteriores e acrescentam
`tool_version`, `scenario_id`, `repetition` e `publish_pattern`.
O modo `connections` também gera `summary.csv`, com `publish_pattern` vazio,
além de `summary.json`, `connections.csv` e `effective-config.conf`.

Há exemplos para ambos os brokers em `configs/pilot/` e `configs/official/`.
Os exemplos oficiais são modelos para ajustar à matriz experimental aprovada.
Substitua os IPs fictícios antes de executar. Para repetir um cenário, ajuste
`repetition` e o sufixo R do identificador juntos.

Validação local:

```powershell
go test ./... -count=1
go vet ./...
go build -o mqttbench.exe ./cmd/mqttbench
```

Os testes de integração usam um servidor MQTT 5 mínimo em loopback (QoS 0),
verificam 75 mensagens por padrão, offsets no `messages.csv`, overrides,
metadados e proteção da senha. Não substituem os pilotos com EMQX/Mosquitto
e QoS 1/2.