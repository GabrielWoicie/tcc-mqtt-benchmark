# Resultados dos ensaios de mensagens MQTT

**Projeto:** Projeto e avaliação de desempenho de uma plataforma IoT baseada em MQTT com EMQX  
**Data dos ensaios AWS:** 07/11/2026 (UTC)  
**Ferramenta de benchmark:** `mqttbench v0.3.0`, em Go, utilizando MQTT 5.0  
**Brokers:** EMQX Community 5.8.9 e Eclipse Mosquitto 2.1.2  
**Escopo:** matriz oficial de mensagens pela Internet, ensaios exploratórios em rede privada AWS e diagnósticos do EMQX C10000.

Os [diretórios de execução, configurações e métricas](../experimentos-tcc.md) estão organizados por família de ensaio. Os resultados específicos de conexões são apresentados separadamente em [resultados da matriz pela Internet](resultados-benchmark.md) e [ensaios de 50.000 conexões](resultados-conexoes-50k.md).

Os testes oficiais de mensagens foram executados com gerador Windows pela Internet. Os testes de estresse utilizaram gerador Linux na rede privada AWS. As diferenças de rede, janela de drenagem, hardware do gerador e carga exigem análise separada das duas baterias.

## 1. Síntese executiva

1. **Matriz oficial pela Internet (30 execuções de mensagens):** nos cenários QoS 1 de 500 e 1.000 publicadores, o EMQX apresentou maior vazão observada no consumidor remoto do que o Mosquitto, enquanto o Mosquitto utilizou menos CPU e RAM. O QoS 0 atingiu praticamente a taxa alvo para 500 publicadores em ambos; QoS 1/2 reduziram o recebimento observado. Os números oficiais são médias de três repetições por cenário.
2. **Estresse AWS C2500:** ambos obtiveram 100% de recebimento observado ao final da drenagem, com cerca de 2.500 mensagens/s. O Mosquitto teve menor latência típica (P50), e o EMQX menor P99.
3. **Estresse AWS C5000:** Mosquitto recebeu 300.000/300.000 mensagens (100%); EMQX recebeu 296.117/299.996 (98,71%). Mosquitto apresentou menor CPU e menor P99 durante o teste de mensagens.
4. **Estresse AWS C10000:** Mosquitto recebeu 564.206/599.998 mensagens confirmadas (94,03%). EMQX registrou 599.999 publicações confirmadas, mas **zero mensagens contabilizadas da fase de medição** pelo coletor. Os logs posteriores identificaram **fila de mensagens do consumidor saturada** e, na execução original, **proteção por estouro da mailbox**. Esse zero **não comprova descarte definitivo de 100% das publicações pelo broker**.
5. **Diagnóstico do EMQX C10000 sem alterar o `mqttbench`:** duas execuções adicionais ampliaram, primeiro, limites da fila e da mailbox para 10.000 e, depois, `max_inflight` para 256. O resultado permaneceu com recebimento observado igual a zero e houve registro de fila cheia nas duas tentativas.
6. **Comparação nas condições avaliadas:** o Mosquitto apresentou menor consumo de recursos e maior recebimento observado nos ensaios AWS C5000/C10000. O resultado caracteriza a arquitetura com um único consumidor e os recursos utilizados, sem determinar a capacidade máxima interna de cada broker.

## 2. Organização das famílias de ensaios

| Família | Ambiente gerador | Métricas de interesse | Repetições | Status |
|---|---|---|---|---|
| Matriz científica oficial — mensagens | Windows externo → brokers EC2 pela Internet | Entrega, vazão, QoS, recursos | 3 por cenário, 30 execuções no total | Concluída e separada do estresse |
| Diagnósticos oficiais QoS | Windows externo → brokers EC2 pela Internet | Efeito de alongar drain | 4 execuções adicionais | Concluídos, **fora** das 30 oficiais |
| Estresse de mensagens AWS | EC2 APP Linux → IP privado das EC2 brokers | Publicação, recepção, latência, CPU, RAM, rede | 1 por broker e carga: **6 execuções** | Concluído |
| Investigação adicional EMQX C10000 | EC2 APP Linux → EMQX | Efeito de limites internos | **2 execuções adicionais** | Concluída; indicadores e eventos de log nas seções 10 e 11 |

A matriz oficial totalizou 48 execuções, sendo **30 de mensagens e 18 de conexões**. Este relatório reproduz **somente os resultados de mensagens**. Os seis cenários AWS e as duas repetições diagnósticas não integram nem substituem a matriz oficial.

---

# PARTE I — Resultados oficiais de mensagens (Internet)

## 3. Método da matriz oficial

A ferramenta `mqttbench v0.3.0` abriu múltiplos publicadores e um consumidor único, com medições de latência ponta a ponta no gerador remoto. Nos ensaios oficiais: payload 256 bytes; 1 mensagem/s por publicador; publicação escalonada (*staggered*); warmup 10 s; publicação 60 s; drenagem de 5 s. Cada cenário de mensagens utilizou **três repetições**.

- **Experimento A — crescimento da carga:** QoS 1; 100, 500 e 1.000 publicadores; EMQX e Mosquitto; 18 execuções.
- **Experimento B — impacto do QoS:** 500 publicadores; QoS 0, 1 e 2; cenários QoS 1 já incluídos no Experimento A; 12 novas execuções; **30 execuções oficiais de mensagens** no total.

### 3.1 Crescimento de publicadores — QoS 1

Os valores apresentados são médias aritméticas das três repetições de cada condição da matriz oficial. As execuções estão identificadas na [seleção de mensagens pela Internet](../experimentos-tcc.md#internet-mensagens).

| Clientes | Vazão recebida EMQX (msg/s) | Vazão recebida Mosquitto (msg/s) | Recebimento EMQX | Recebimento Mosquitto |
|---:|---:|---:|---:|---:|
| 100 | 100,00 | 100,00 | 100,00% | 100,00% |
| 500 | 216,48 | 129,13 | 43,30% | 25,84% |
| 1.000 | 220,39 | 130,48 | 22,04% | 13,05% |

**Interpretação:** a partir de C500/Q1, o recebimento registrado pelo consumidor remoto não acompanhou a taxa solicitada; nos cenários citados, o EMQX apresentou maior vazão observada. Não significa que o broker tenha recusado todas as mensagens não observadas. O gerador e o consumidor pela Internet fazem parte do sistema medido.

### 3.2 Impacto do nível de QoS — 500 clientes

| QoS | Vazão EMQX (msg/s) | Vazão Mosquitto (msg/s) | Recebimento EMQX | Recebimento Mosquitto |
|---:|---:|---:|---:|---:|
| 0 | 499,67 | 500,00 | 100,00% | 100,00% |
| 1 | 216,48 | 129,13 | 43,30% | 25,84% |
| 2 | 109,09 | 55,83 | 21,82% | 11,17% |

**Interpretação:** no ambiente remoto, o aumento dos requisitos de confirmação de QoS 0 para QoS 1 e QoS 2 coincidiu com menor recebimento observado, principalmente na conexão única do consumidor. Os valores de QoS 1 acima reutilizam os mesmos cenários do Experimento A; não são novas execuções.

### 3.3 Consumo do processo broker — matriz oficial, QoS 1

| Clientes | RAM EMQX (MB) | RAM Mosquitto (MB) | CPU média EMQX | CPU média Mosquitto |
|---:|---:|---:|---:|---:|
| 100 | 259,50 | 9,58 | 5,34% | 1,21% |
| 500 | 283,24 | 12,20 | 13,93% | 2,67% |
| 1.000 | 320,29 | 14,62 | 25,74% | 4,61% |

A memória da matriz oficial é apresentada em MB. A CPU refere-se ao processo do broker. As tabelas dos ensaios AWS usam MiB para memória; as unidades são mantidas em cada tabela.

## 4. Diagnósticos QoS da matriz oficial (fora das 30 execuções)

Foi testada a hipótese de que a redução do recebimento em QoS 1/2 seria explicada apenas por mensagens pendentes que precisavam de maior tempo para escoar. Com 500 clientes e 60 s de publicação, o `drain` foi aumentado para 180 s (QoS 1) e 420 s (QoS 2).

| Broker / cenário | Publicadas | Recebidas | Recebimento observado | Vazão normalizada de recebimento |
|---|---:|---:|---:|---:|
| EMQX C500/Q1 | 30.000 | 13.000 | 43,33% | 216,67 msg/s |
| Mosquitto C500/Q1 | 30.000 | 8.144 | 27,15% | 135,73 msg/s |
| EMQX C500/Q2 | 30.000 | 7.016 | 23,39% | 116,93 msg/s |
| Mosquitto C500/Q2 | 30.000 | 4.021 | 13,40% | 67,02 msg/s |

O aumento de drenagem não restaurou o recebimento para 100%. Esses quatro testes integram os diagnósticos pela Internet. Nos cenários AWS, a rede e o gerador foram modificados e a drenagem foi definida em 60 s.

---

# PARTE II — Experimentos extremos de mensagens na AWS (rede privada)

## 5. Infraestrutura e procedimento

- **Broker EMQX:** EC2 com 2 vCPUs e aproximadamente 2 GiB de RAM, EMQX Community 5.8.9; conexão MQTT em `172.31.12.115:1883`.
- **Broker Mosquitto:** EC2 de recursos equivalentes, Eclipse Mosquitto 2.1.2; conexão MQTT em `172.31.13.183:1883`.
- **Gerador:** EC2 APP Linux na mesma VPC (`172.31.2.214`), com `mqttbench v0.3.0`. O cenário C2500 ocorreu com aproximadamente 2 GiB de RAM no gerador; **antes de C5000/C10000 o gerador foi redimensionado para 8 GiB** por causa de OOM na execução de C5000 com hardware anterior.
- **Broker testado separadamente:** mesmos parâmetros de publicação e consumidor para cada produto; backend IoT, frontend e PostgreSQL não integram o fluxo de medição.
- **Cliente consumidor:** uma única conexão MQTT dedicada que assina o prefixo exclusivo do cenário em `tcc/benchmark/stress/<run-id>/#` e recebe QoS 1.
- **Invariantes:** MQTT 5.0, QoS 1, payload configurado de 256 bytes, 1 mensagem/s/publicador, taxa de abertura configurada de 100 conexões/s, publicação *staggered*, warmup de 10 s, duração de medição de 60 s, drenagem de 60 s, TCP/1883 sem TLS (conforme configurações efetivas).
- **Identificação das execuções:** cada teste possui um `run_id` distinto. Os resumos, as configurações efetivas e a telemetria estão vinculados na [tabela de mensagens AWS](../experimentos-tcc.md#aws-mensagens).

**Atenção aos tempos:** abrir 10.000 conexões a uma taxa solicitada de 100/s leva aproximadamente 100 s, antes do warmup. Portanto, um defeito ou desconexão do único consumidor na preparação pode comprometer toda a medição. O `mqttbench` utiliza `CleanStart=true` e `SessionExpiryInterval=0` no coletor; a sessão não é persistida após desconexão.

## 6. Resultado primário — conexão, publicação e recebimento

| Publicadores | Broker | Conectados | Pub. confirmadas | Recebidas únicas | Não observadas | Recebimento observado |
|---:|---|---:|---:|---:|---:|---:|
| 2.500 | EMQX | 2.500 | 150.000 | 150.000 | 0 | 100,00% |
| 2.500 | Mosquitto | 2.500 | 149.998 | 149.998 | 0 | 100,00% |
| 5.000 | EMQX | 5.000 | 299.996 | 296.117 | 3.879 | 98,71% |
| 5.000 | Mosquitto | 5.000 | 300.000 | 300.000 | 0 | 100,00% |
| 10.000 | EMQX | 10.000 | 599.999 | 0 | 599.999 | 0,00% |
| 10.000 | Mosquitto | 10.000 | 599.998 | 564.206 | 35.792 | 94,03% |

Em todas as seis execuções originais:

- `publisher_connect_failures=0` e `publish_errors=0`;
- o número de publicadores conectados corresponde ao solicitado;
- os campos `duplicates` e `decode_errors` são zero;
- **`publish_success` indica sucesso da chamada Publish QoS 1 / confirmação pelo broker**; não confirma entrega ao consumidor final;
- **`lost` corresponde a `publish_success − received_unique`** ao encerramento e representa mensagens não observadas na janela de coleta. Esse indicador não identifica perdas definitivas no broker.

**Vazão normalizada de recebimento**, calculada pela ferramenta como `received_unique / duration_seconds`:

| Carga | EMQX (msg/s) | Mosquitto (msg/s) |
|---:|---:|---:|
| C2500 | 2.500,00 | 2.499,97 |
| C5000 | 4.935,28 | 5.000,00 |
| C10000 | 0,00 contabilizada | 9.403,43 |

**Nota:** como o contador de recebimento pode incorporar mensagens entregues durante o período de drenagem posterior, dividir por 60 s é uma **normalização definida pelo programa**, não necessariamente a taxa real instantânea de recepção em cada segundo.

## 7. Latência das mensagens recebidas

| Publicadores | Broker | Média (ms) | P50 (ms) | P95 (ms) | P99 (ms) | Máximo (ms) | Amostras |
|---:|---|---:|---:|---:|---:|---:|---:|
| 2.500 | EMQX | 0,629 | 0,537 | 0,869 | 3,034 | 20,238 | 150.000 |
| 2.500 | Mosquitto | 0,594 | 0,418 | 0,761 | 4,634 | 44,452 | 149.998 |
| 5.000 | EMQX | 3,348 | 0,491 | 1,333 | 132,899 | 477,105 | 296.117 |
| 5.000 | Mosquitto | 0,842 | 0,370 | 0,743 | 17,133 | 79,434 | 300.000 |
| 10.000 | EMQX | N/D | N/D | N/D | N/D | N/D | 0 |
| 10.000 | Mosquitto | 10,238 | 0,397 | 10,608 | 244,600 | 1.418,717 | 564.206 |

A latência é calculada no gerador entre carimbo de envio e recebimento para as mensagens da fase `measure` que passaram pelos filtros do consumidor. Os percentis se referem **somente às mensagens observadas**.

- **C2500:** Mosquitto apresentou menor P50, mas EMQX apresentou melhor P99 e menor máximo. Ambos tiveram P99 inferior a 5 ms.
- **C5000:** P99 foi 132,899 ms no EMQX contra 17,133 ms no Mosquitto; o EMQX também teve menor taxa observada de recebimento.
- **C10000:** Mosquitto teve P99 de 244,600 ms para as mensagens observadas e máximo de cerca de 1,419 s. O EMQX não possui amostras de latência da fase medida; por isso, a tabela apresenta **N/D**. O zero no JSON corresponde à ausência de amostras, sem representar latência de 0 ms.
- No C5000, **3.571** mensagens recebidas do EMQX tiveram latência superior a 100 ms, enquanto nenhuma mensagem recebida do Mosquitto excedeu esse limiar. No C10000 Mosquitto, **11.385** mensagens observadas excederam 100 ms. Esses totais referem-se apenas ao subconjunto recebido.

### 7.1 Latência de conexão dos publicadores (métrica acessória do modo message)

| Publicadores | Broker | P99 conexão dos publicadores | Falhas de conexão | Erros de publicação | Duplicatas | Erros de decodificação |
|---:|---|---:|---:|---:|---:|---:|
| 2.500 | EMQX | 1,991 ms | 0 | 0 | 0 | 0 |
| 2.500 | Mosquitto | 1,219 ms | 0 | 0 | 0 | 0 |
| 5.000 | EMQX | 8,942 ms | 0 | 0 | 0 | 0 |
| 5.000 | Mosquitto | 3,273 ms | 0 | 0 | 0 | 0 |
| 10.000 | EMQX | 8,527 ms | 0 | 0 | 0 | 0 |
| 10.000 | Mosquitto | 2,050 ms | 0 | 0 | 0 | 0 |

A latência CONNECT desta seção não substitui as séries e diagnósticos dos experimentos específicos `mode=connections`.

## 8. CPU, memória, conexões TCP e rede dos brokers

As métricas foram coletadas nas EC2 dos brokers com `sar`, `pidstat` e `ss`. As médias de CPU referem-se à janela aproximada de publicação, excluindo as demais fases da coleta. A CPU de processo no `pidstat` admite valores acima de 100% quando utiliza mais de um vCPU; a CPU da EC2 é o percentual agregado da instância. RSS representa a memória residente total amostrada do processo, incluindo a linha de base.

| Publicadores | Broker | CPU média processo na publicação | CPU média EC2 na publicação | Pico RSS processo (MiB) | Menor RAM disponível EC2 (MiB) | Pico TCP EC2 |
|---:|---|---:|---:|---:|---:|---:|
| 2.500 | EMQX | 63,8% | 37,0% | 375,4 | 1.137,3 | 2.501 |
| 2.500 | Mosquitto | 14,0% | 12,5% | 252,3 | 1.226,6 | 2.501 |
| 5.000 | EMQX | 126,8% | 70,3% | 524,3 | 1.044,8 | 5.001 |
| 5.000 | Mosquitto | 27,9% | 25,5% | 33,6 | 1.494,1 | 5.001 |
| 10.000 | EMQX | 127,5% | 72,1% | 675,9 | 879,0 | 10.000 |
| 10.000 | Mosquitto | 52,7% | 43,4% | 58,4 | 1.443,8 | 10.001 |

**Cuidados de comparação:**

1. No **C5000**, o Mosquitto teve menor CPU de processo e menor P99 e registrou mais mensagens no consumidor que o EMQX; é o cenário comparativo mais informativo entre os testes extremos com encaminhamento efetivamente observado em ambos.
2. No **C10000**, a quantidade de mensagens efetivamente observadas foi muito diferente entre os brokers; os consumos de CPU correspondem a volumes distintos de entregas.
3. Entre as baterias C2500 e C5000, o estado dos processos mudou, incluindo reinicializações; o RSS do Mosquitto era ~252 MiB na bateria anterior e ~34–58 MiB na posterior. A variação de RSS entre essas baterias não permite isolar o efeito do aumento da carga.
4. A EC2 APP com 8 GiB foi dimensionada para evitar o OOM do gerador. Não há séries completas de CPU/RAM do gerador versionadas para C5000 e C10000; a conclusão dos testes não demonstra ausência de outros gargalos.

### 8.1 Rede na fase de publicação

| Publicadores | Broker | RX médio da EC2 (kB/s) | TX médio da EC2 (kB/s) |
|---:|---|---:|---:|
| 2.500 | EMQX | 1.640 | 1.378 |
| 2.500 | Mosquitto | 1.628 | 1.362 |
| 5.000 | EMQX | 3.277 | 2.672 |
| 5.000 | Mosquitto | 3.243 | 2.652 |
| 10.000 | EMQX | 5.210 | 1.402 |
| 10.000 | Mosquitto | 6.264 | 4.898 |

Esses números são da interface de rede como um todo, com overhead e eventuais outros pacotes; não equivalem diretamente às contagens de PUBLISH encaminhadas. O TX inferior do EMQX C10000 é compatível com o problema observado na entrega, mas não é prova isolada da causa.

---

# PARTE III — Diagnóstico do cenário EMQX C10000

## 9. Evidências da execução original (sem ajustes)

**Run ID:** `20261107T135159.299404144Z` (identificador seguro da pasta `20261107T135159299404144Z`).  
**Início UTC:** 07/11/2026, às 13:51:59.  
**Registro da execução:** [summary.json](../../results/STRESS-MSG-EMQX-C10000-Q1-AWS-R1_emqx_20261107T135159299404144Z/summary.json).  
**Resultado:** 10.000 publicadores conectados; 599.999 publicações confirmadas; 0 mensagens da fase medida contabilizadas pelo consumidor.

**Log EMQX — 13:53:41 UTC:**

```text
[warning] clientid: tcc-collector-59299404144Z,
msg: dropped_msg_due_to_mqueue_is_full,
queue: #{len => 1000,dropped => 1,max_len => 1000}

[warning] clientid: tcc-collector-59299404144Z,
msg: check_oom_shutdown,
shutdown: #{max => 1000,reason => mailbox_overflow,value => 1484}
```

**Diagnóstico amparado pelo log:** a fila da sessão do consumidor atingiu o limite de 1.000 e a proteção contra excesso de mensagens na mailbox do processo do EMQX foi acionada (1.484 > 1.000), encerrando a conexão do consumidor. Esse mecanismo interno difere do OOM Killer do Linux que encerrou o processo do broker no ensaio de 50.000 conexões.

O monitor TCP observou máximo de 10.000 conexões (quando eram esperadas 10.001 contando publicadores + coletor); trata-se de indício adicional, mas **a série amostral de TCP não identifica individualmente o cliente ausente**.

## 10. Reteste 1 — fila e mailbox ampliadas (mesmo executável)

**Run ID:** `20261107T155303.390657397Z` (15:53–15:56 UTC).  
**Configuração do broker:** `mqtt.max_mqueue_len=10000`, `force_shutdown.max_mailbox_size=10000`, `mqtt.max_inflight=32` e proteção ativa; a leitura posterior de `max_heap_size` indicou `32MB`.  
**Resultado:** 10.000 publicadores conectados, 599.999 publicações bem-sucedidas, 0 mensagens da fase medida contabilizadas, nenhuma falha de publicação.

**Log EMQX:**

```text
2026-11-07T15:54:47.018647+00:00
clientid: tcc-collector-03390657397Z
msg: dropped_msg_due_to_mqueue_is_full
queue: #{len => 10000,dropped => 1,max_len => 10000}

2026-11-07T15:55:17.863907+00:00
msg: log_events_throttled_during_last_period
dropped: #{dropped_msg_due_to_mqueue_is_full => 21318}
```

**Interpretação:** a fila voltou a atingir seu limite e ocorreram descartes; ampliar a capacidade de 1.000 para 10.000 não restaurou o recebimento observado na fase medida. O número `21318` corresponde a eventos de log suprimidos por throttling; esse contador não determina a quantidade exata de mensagens QoS 1 definitivamente perdidas.

## 11. Reteste 2 — inflight também ampliado (mesmo executável)

**Run ID:** `20261107T161242.481842647Z` (16:12–16:16 UTC).  
**Configuração do broker:**

```hocon
mqtt.max_inflight = 256
mqtt.max_mqueue_len = 10000
force_shutdown.max_mailbox_size = 10000
force_shutdown.enable = true
force_shutdown.max_heap_size = "32MB"
```

**Resultado:** 10.000 publicadores conectados; 599.999/599.999 publicações bem-sucedidas; 0 mensagens da fase medida contabilizadas; 0 duplicatas; 0 erros de decodificação; 0 erros de publicação.

**Logs EMQX:**

```text
2026-11-07T16:14:24.863639+00:00
clientid: tcc-collector-42481842647Z
msg: dropped_msg_due_to_mqueue_is_full
queue: #{len => 10000,dropped => 1,max_len => 10000}

2026-11-07T16:14:26.300270+00:00
clientid: tcc-collector-42481842647Z
msg: dropped_msg_due_to_mqueue_is_full
queue: #{len => 10000,dropped => 8634,max_len => 10000}

2026-11-07T16:14:26.273606+00:00
log_events_throttled_during_last_period:
dropped_msg_due_to_mqueue_is_full => 8632

2026-11-07T16:15:26.275449+00:00
log_events_throttled_during_last_period:
dropped_msg_due_to_mqueue_is_full => 12602
```

**Interpretação:** o terceiro teste voltou a saturar a fila de 10.000 mensagens. Os eventos apresentados não exibem novo `mailbox_overflow`, diferentemente da primeira execução. Esses registros não permitem determinar se a sessão permaneceu conectada até o final. Os contadores de mensagens do `mqttbench` continuaram em zero.

### 11.1 Consolidação dos três C10000 EMQX

| Execução (identificada pelo `run_id`) | Fila máxima | Mailbox máxima | Inflight máximo | Publicações confirmadas | Recebidas da fase medida | Evidência nos logs |
|---|---:|---:|---:|---:|---:|---|
| Original — `135159299404144Z` | 1.000 | 1.000 | 32 | 599.999 | 0 | Fila cheia + proteção `mailbox_overflow` |
| Ajuste 1 — `155303390657397Z` | 10.000 | 10.000 | 32 | 599.999 | 0 | Fila cheia, log de descartes suprimidos |
| Ajuste 2 — `161242481842647Z` | 10.000 | 10.000 | 256 | 599.999 | 0 | Fila cheia, log de descartes suprimidos |

**Identificação dos retestes:** ambos reutilizaram `scenario_id=STRESS-MSG-EMQX-C10000-Q1-AWS-R1` e `repetition=1`, mas possuem `run_id` e configurações distintos. Os indicadores e eventos de log das seções 10 e 11 documentam esses ensaios. Os respectivos arquivos individuais `summary.json` e séries de CPU/rede não estão versionados; a telemetria da execução original não representa os retestes.

## 12. Alcance do indicador de recebimento

No `mqttbench v0.3.0`, o método `onPublish` ignora mensagens cujo `run_id` não coincide com a execução ou cujo campo `phase` não seja `measure`:

```go
if p.RunID != c.runID || p.Phase != "measure" {
    return true, nil
}
```

Portanto, o contador da ferramenta só cobre mensagens da fase de medição. Logs do broker com `phase:"warmup"` indicam que o congestionamento já ocorria na preparação e, na primeira execução, havia proteção ativada antes da janela principal.

Além disso, QoS 1 confirma a transferência **publicador → broker** via PUBACK, mas não prova que o PUBLISH foi entregue e confirmado pelo assinante. A arquitetura de consumidor único pode ser limitada por vazão de rede, controle de fluxo, processamento do callback ou política de gerenciamento de filas do broker. **Os logs comprovam descartes por fila cheia, mas não quantificam o destino individual das 599.999 publicações da janela de medição.**

## 13. Políticas de fila: parâmetros observados e padrões documentados

| Parâmetro | EMQX original/observado | EMQX ajustado (última tentativa) | Mosquitto 2.1.2 |
|---|---|---|---|
| Fila da sessão | `mqtt.max_mqueue_len=1000` | `10000` | Sem override em `/etc/mosquitto`: `max_queued_messages` padrão **1.000** |
| Janela de envios pendentes | `mqtt.max_inflight=32` | `256` | Sem override: `max_inflight_messages` padrão **20** |
| Mailbox interna de processo | Proteção por `max_mailbox_size=1000` | `10000` | Não se aplica: implementação em C, sem mailbox Erlang correspondente |
| Memória | Proteções internas, com `force_shutdown.max_heap_size` (valor atual confirmado `32MB`) | `32MB` na última leitura | `memory_limit` não definido explicitamente, padrão sem limite fixo do broker |
| `max_queued_bytes` | N/A — parâmetro não equivalente direto | N/A | Sem override; padrão `0` (sem limite adicional por bytes) |

**Verificação das configurações Mosquitto:** `systemctl cat mosquitto` mostrou `ExecStart=/usr/sbin/mosquitto -c /etc/mosquitto/mosquitto.conf`, com `include_dir /etc/mosquitto/conf.d`. A pesquisa `grep -R` nos arquivos `/etc/mosquitto/` não encontrou definições para `max_inflight_messages`, `max_queued_messages`, `max_queued_bytes` ou `memory_limit`. Portanto, os limites do Mosquitto nessa tabela são padrões documentados, inferidos da ausência de overrides, sem leitura direta dos valores internos em tempo de execução. O serviço Mosquitto também tinha `LimitNOFILE=1048576:1048576`, referente ao limite de descritores abertos.

**Referências técnicas dos parâmetros:**

- Mosquitto, `mosquitto-conf(5)`: https://mosquitto.org/man/mosquitto-conf-5.html
- EMQX, *Inflight Window and Message Queue*: https://docs.emqx.com/en/emqx/latest/develop/design/inflight-window-and-message-queue.html

**Nota:** os dois brokers possuem mecanismos de fila por sessão, mas o limite numérico não determina isoladamente o throughput; o tratamento de mensagens em trânsito, o comportamento do cliente, a implementação e as configurações de segurança e entrega também importam.

---

# PARTE IV — Limitações e discussão

## 14. Limitações e cuidados interpretativos

1. **Uma execução por condição extrema principal:** os seis testes AWS têm somente uma execução por broker/carga; não há variabilidade amostral suficiente para alegações de significância estatística. Os dois C10000 adicionais são tentativas com **configuração alterada**, não repetições homogêneas do cenário original.
2. **Mudança de hardware do gerador:** C2500 foi medido com ~2 GiB de RAM no gerador; C5000/C10000 com 8 GiB. Além da carga, houve mudança nos recursos do gerador entre essas baterias.
3. **Ausência de telemetria completa do gerador:** após o redimensionamento, estão versionadas as métricas dos brokers, sem um perfil completo de CPU/RAM do gerador. A conclusão das execuções demonstra que esses testes foram completados, mas não exclui limitações de geração ou consumo de mensagens.
4. **Consumidor único:** a assinatura, o protocolo QoS, a janela inflight e as filas por sessão podem influenciar o escoamento, assim como o processamento interno do gerador. Essa arquitetura não mede isoladamente o limite máximo de publicação do broker.
5. **SUBACK e observabilidade:** a versão `mqttbench v0.3.0` não registra adequadamente eventos assíncronos do coletor nem discrimina em detalhe mensagens brutas recebidas na fase `warmup`, mensagens filtradas e desconexões. Em C10000 EMQX houve ausência de mensagens contabilizadas na fase `measure`, com fila saturada confirmada.
6. **PUBACK não é entrega fim a fim:** uma publicação QoS 1 confirmada pelo broker pode não ser recebida pelo consumidor que subscreveu o tópico. A combinação `publish_success=599999` e `received_unique=0` não determina o destino individual das 599.999 publicações.
7. **Drenagem finita:** no estresse AWS, a ferramenta aguarda 60 s após o envio antes de consolidar o resultado; entregas posteriores não aparecem no relatório.
8. **Mudança de parâmetros do EMQX:** os retestes alteraram `mqtt.max_mqueue_len`, `force_shutdown.max_mailbox_size` e `mqtt.max_inflight`. Integram a investigação diagnóstica e não substituem o EMQX original na comparação direta com Mosquitto.
9. **RSS inicial variável:** a memória do processo inclui a linha de base; os brokers foram reiniciados ou estavam em estados diferentes entre baterias. As medidas de RSS não representam diretamente o custo incremental por mensagem ou cliente.
10. **Sem TLS e autenticação nos extremos AWS:** os resultados se restringem às conexões sem TLS e autenticação adotadas nos ensaios. Não avaliam ACL complexa, múltiplos assinantes ou cluster.
11. **Cenários distintos têm conclusões distintas:** a matriz oficial por Internet favoreceu o EMQX em vazão de recebimento para QoS 1/2 em cargas específicas; os testes extremos de rede privada favoreceram o Mosquitto em C5000/C10000. As observações **não se anulam**: as condições experimentais mudaram materialmente.
12. **OOM anterior do gerador:** as primeiras tentativas de C5000 com APP ~2 GiB foram encerradas pelo kernel por falta de RAM no gerador. Essas tentativas ficaram fora da comparação entre os brokers. Depois do redimensionamento, os cenários foram concluídos.

## 15. Discussão conjunta e conclusão científica

Os dados demonstram **dois perfis distintos**. Na matriz oficial executada com gerador remoto pela Internet, o EMQX recebeu mais mensagens no consumidor em cenários de QoS 1 e QoS 2 a partir de 500 publicadores, enquanto o Mosquitto foi mais econômico em CPU e RAM. Já nos testes exploratórios extremos em VPC AWS, ambos atingiram 100% de recebimento observado em C2500, mas o Mosquitto manteve melhor resultado em C5000 e C10000: entrega integral no C5000 e 94,03% no C10000. O EMQX apresentou 98,71% no C5000 e saturação comprovada da fila do consumidor único no C10000, com zero mensagens contabilizadas na fase de medição. O aumento da fila, da mailbox e da janela inflight não eliminou essa saturação.

Nas condições testadas, o Mosquitto apresentou menor consumo de recursos e maior recebimento observado nas maiores cargas AWS. O EMQX apresentou saturação da fila do consumidor único no cenário com 10.000 publicadores, inclusive após a ampliação dos limites de fila, mailbox e inflight. A arquitetura, o hardware e as diferenças entre as baterias delimitam essa comparação. A quantidade de mensagens não observadas e a ausência de repetições homogêneas nos ensaios extremos impedem quantificar perdas definitivas ou estabelecer superioridade estatística entre os brokers.
