# Resultados dos ensaios MQTT de 50.000 conexões na AWS

**Data das execuções:** 7 de novembro de 2026 (UTC)  
**Ferramenta:** `tcc-mqtt-benchmark` (`mqttbench v0.3.0`)  
**Brokers:** EMQX e Eclipse Mosquitto  
**Natureza:** experimentos complementares/exploratórios, independentes da matriz oficial do TCC.

Os [quatro ensaios de 50.000 conexões](../experimentos-tcc.md#aws-50k) estão identificados por cenário, horário UTC e diretório de execução. Os resumos, as configurações efetivas e as métricas dos brokers e do gerador estão vinculados no índice.

## 1. Síntese dos resultados

| Teste | Broker | Taxa configurada (tentativas/s) | Clientes solicitados | MQTT conectados segundo o mqttbench | Falhas | Taxa de sucesso | P99 CONNECT | Taxa efetiva (conexões/s) |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| R4 | EMQX | 1.000 | 50.000 | 50.000 | 0 | 100% | 187,89 ms | 447,49 |
| R4 | Mosquitto | 1.000 | 50.000 | 50.000 | 0 | 100% | 158,52 ms | 417,48 |
| R5 | EMQX | 0 (sem limitador) | 50.000 | 31.263 | 18.737 | 62,53% | 15.436,71 ms | 579,70 |
| R5 | Mosquitto | 0 (sem limitador) | 50.000 | 31.784 | 18.216 | 63,57% | 10.808,48 ms | 422,58 |

**Importante:** `clients_connected` é o número de conexões que chegaram a ter sucesso no handshake MQTT; **não atesta que todas permaneceram vivas até o fim do hold**. A taxa efetiva é `clients_connected / connection_window_seconds`, por isso depende da duração da janela, de atrasos e timeouts; não é uma medida isolada da capacidade máxima do broker.

A ampliação do intervalo de portas efêmeras do gerador eliminou as falhas por `cannot assign requested address` observadas nos testes anteriores. Sob 1.000 tentativas/s, ambos conseguiram 50.000 conexões. Sem limitador, houve falhas significativas.

## 2. Evidência conclusiva de OOM no EMQX — R5

No dia **07/11/2026, às 03:45:00 UTC**, o kernel Linux registrou:

```text
erts_sched_1 invoked oom-killer
... task_memcg=/system.slice/emqx.service,task=beam.smp,pid=559 ...
Out of memory: Killed process 559 (beam.smp) ... anon-rss:1443100kB ... shmem-rss:88424kB
```

O systemd confirmou:

```text
emqx.service: A process of this unit has been killed by the OOM killer.
emqx.service: Main process exited, code=killed, status=9/KILL
emqx.service: Failed with result 'oom-kill'.
emqx.service: Consumed ... 1.4G memory peak, 0B memory swap peak.
```

A sequência temporal cruzada com os arquivos de monitoramento foi:

- **03:44:08 UTC:** início do teste R5 do EMQX.
- **03:44:38 UTC:** amostragem do servidor identificou pico de **33.526 conexões TCP** na porta MQTT.
- **03:45:00 UTC:** OOM Killer encerrou `beam.smp`, após forte pressão de memória. O processo atingiu aproximadamente **1.499 MiB RSS** nas amostras; a memória disponível da EC2 chegou a cerca de **25 MiB** nas métricas previamente coletadas do broker.
- **03:45:07 UTC:** a amostra de conexões no gerador já registra **zero** conexões TCP com o EMQX.
- **03:46:02 UTC:** encerramento do `mqttbench`.

**Conclusão sobre R5 do EMQX:** houve indisponibilidade comprovada do serviço MQTT causada pelo esgotamento de memória do **servidor do broker**, e não apenas falhas de conexão do gerador. O OOM ocorreu aproximadamente no final da fase de abertura, antes do período integral de manutenção. Portanto, as 31.263 conexões contabilizadas como bem-sucedidas não foram sustentadas durante os 60 segundos completos.

No R5 do Mosquitto, o processo continuou presente nas métricas coletadas; a contagem TCP estabilizou em **31.784** durante a fase de manutenção, antes do encerramento normal. Não há evidência de queda equivalente do processo Mosquitto nessa execução.

## 3. Métricas da EC2 geradora

A coleta do gerador está organizada em um diretório por execução em [`generator-metrics/`](../../generator-metrics/), com os arquivos `sistema.txt`, `processo.txt`, `conexoes.csv` e `oom.txt`.

As medidas abaixo consideram apenas o período indicado em `started_at_utc` e `finished_at_utc` de cada `summary.json` (o monitoramento EMQX R4 permaneceu ativo muito depois de o teste acabar, e esse trecho ocioso foi excluído).

| Métrica do gerador | EMQX R4 | Mosquitto R4 | EMQX R5 | Mosquitto R5 |
|---|---:|---:|---:|---:|
| Pico de CPU da EC2 APP | 100% | 100% | 100% | 100% |
| Pico de CPU do processo `mqttbench_linux` | 200% | 199,5% | 200% | 200% |
| Pico RSS do processo | 1.012 MiB | 918 MiB | 1.246 MiB | 916 MiB |
| Menor RAM disponível na APP | 297 MiB | 316 MiB | 354 MiB | 647 MiB |
| Pico TCP no gerador (`ss`, amostrado a cada ~10 s) | 50.000 | 50.000 | 33.394 | 35.977 |
| OOM registrado nos arquivos da APP | Não | Não | Não | Não |

**Interpretação:** o gerador trabalhou no limite dos seus **dois vCPUs** em todos os cenários, especialmente na abertura das conexões, o que pode ter influenciado a taxa efetiva e as falhas, principalmente no modo sem limitador. Apesar da pressão sobre CPU e RAM, **não há evidência de OOM na APP** nos arquivos coletados; ela ainda tinha memória disponível. O registro de OOM conclusivo refere-se à EC2 do **EMQX**.

O pico de TCP não equivale necessariamente à quantidade de clientes com handshake MQTT concluído: conexões TCP podem estar estabelecidas enquanto a negociação MQTT está pendente. As contagens são amostrais, não um registro contínuo de cada abertura/fechamento.

## 4. Uso de recursos nos servidores dos brokers

| Métrica observada | EMQX R4 | Mosquitto R4 | EMQX R5 | Mosquitto R5 |
|---|---:|---:|---:|---:|
| Pico RSS do processo broker | ~916 MiB | ~252 MiB | ~1.499 MiB | ~252 MiB |
| Pico de CPU ocupada na EC2 | ~100% | ~52% | ~100% | ~36% |
| Menor memória disponível da EC2 | ~448 MiB | ~1.102 MiB | ~25 MiB | ~1.142 MiB |
| Estado do broker ao longo do teste | Operacional | Operacional | **Encerrado por OOM** | Operacional |

O Mosquitto utilizou substancialmente menos memória e CPU na mesma categoria de hardware e no perfil específico de conexões sem publicação contínua. Os experimentos não determinam o limite universal de cada software nem comprovam o comportamento de aplicações com múltiplas assinaturas, mensagens QoS 1/2 e tráfego constante.

## 5. Interpretações válidas e limites metodológicos

1. **Com taxa controlada de 1.000 conexões/s:** os dois brokers estabeleceram 50.000 conexões sem erros, embora a taxa efetiva tenha sido menor que a configurada; o gerador também chegou a ocupar 100% de CPU.
2. **Com rajada sem limitador:** ambos apresentaram falhas; o EMQX comprovadamente sofreu OOM e interrompeu o serviço; o Mosquitto permaneceu operacional, porém completou apenas 31.784 handshakes MQTT das 50.000 tentativas. Parte das falhas do Mosquitto também pode refletir limites do gerador/filas/sistema operacional, pois sua CPU chegou a 100%.
3. **OOM e eficiência de recursos:** a ocorrência de OOM no EMQX é um resultado experimental **válido sob o hardware fixado**, não uma falha de metodologia. O limite é da combinação broker + configuração + hardware + perfil de carga, e não uma característica universal do broker.
4. **Repetibilidade:** R4 e R5 são uma execução de cada configuração por broker após alteração de portas; as múltiplas repetições anteriores pertencem a condições diferentes. Não há repetições homogêneas suficientes para estabelecer significância estatística nas taxas efetivas ou na diferença de 521 handshakes MQTT no R5.
5. **Metadado:** nos quatro `summary.json`, `repetition` está como `1`, embora os identificadores dos cenários terminem em `R4` ou `R5`. A identificação de cada ensaio usa o `scenario_id`, o horário UTC e o diretório, mantendo os metadados da execução.
6. **Estrutura do teste:** as conexões ficam praticamente ociosas durante o hold. O ensaio avalia abertura e manutenção de conexões, sem determinar o desempenho máximo de encaminhamento de mensagens MQTT.
