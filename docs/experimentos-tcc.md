# Experimentos e resultados do TCC

Este diretório reúne a documentação dos experimentos do TCC: método,
resultados, diagnósticos e limitações. As tabelas abaixo identificam as
execuções que compõem cada análise e os respectivos arquivos no repositório.
A unidade de identificação é **pasta + horário UTC**; `scenario_id` e
`repetition` foram reutilizados em diagnósticos e não identificam sozinhos
uma execução.

As análises detalhadas estão organizadas em três documentos:

- [Resultados do benchmark pela Internet](relatorios/resultados-benchmark.md): método e resultados das 48 execuções oficiais e dos diagnósticos QoS.
- [Resultados de mensagens MQTT](relatorios/resultados-mensagens.md): comparação entre as baterias pela Internet e na rede privada AWS, com os diagnósticos do EMQX C10000.
- [Resultados de 50.000 conexões](relatorios/resultados-conexoes-50k.md): abertura e manutenção de conexões, consumo de recursos e ocorrência de OOM.

## Como localizar a evidência de uma tabela

| Resultado | Seleção de execuções | Indicadores e análise |
|---|---|---|
| Crescimento da carga, QoS 1 (C100/C500/C1000) | [Experimento A, R1–R3 de cada broker](#internet-mensagens) | `receive_throughput_msg_s`, `delivery_rate_pct` e `latency` dos resumos |
| Impacto do QoS (C500, Q0/Q1/Q2) | [Experimento B, R1–R3; Q1 reutiliza A](#internet-mensagens) | As mesmas execuções C500/Q1 entram nas duas análises, sem duplicar a amostra |
| Conexões C500/C1000/C2500 | [Experimento C, R1–R3; somente C2500 corrigido](#internet-conexoes) | `clients_connected`, `connection_latency` e `actual_successful_connect_rate_per_s` |
| CPU/RAM/rede da matriz pela Internet | Mesmas 48 execuções oficiais | CPU/RAM nas seções 9 e 11 dos [resultados do benchmark](relatorios/resultados-benchmark.md); as séries brutas de recursos dessa matriz não estão versionadas |
| Mensagens AWS, C2500/C5000/C10000 | [Seis execuções, uma por broker/carga](#aws-mensagens) | Resultado, configuração efetiva e telemetria do broker |
| Retestes EMQX C10000 | [Dois `run_id` adicionais, configurações alteradas](#emqx-retestes) | Indicadores, configurações e eventos de log nas seções 10 e 11 dos [resultados de mensagens](relatorios/resultados-mensagens.md) |
| 50.000 conexões AWS | [Quatro execuções R4/R5](#aws-50k) | Resultados, métricas do broker e do gerador; [análise dos ensaios](relatorios/resultados-conexoes-50k.md) |

## Matriz oficial pela Internet

Gerador Windows externo aos brokers EC2; MQTT 5.0, ferramenta 0.3.0. São
**30 execuções de mensagens + 18 de conexões = 48**, com três repetições por
condição. Pilotos, C2500 anteriores à correção, diagnósticos QoS e estresse
ficam fora dessas médias.

Nos ensaios de mensagens: 1 msg/s por publicador, payload configurado de 256 B,
publicação `staggered`, warmup 10 s, medição 60 s e drain 5 s. Nas conexões:
taxa solicitada 100/s e hold 60 s. Cada linha vincula os arquivos da execução,
incluindo o resumo e a configuração efetiva.

<a id="internet-mensagens"></a>

### Mensagens: execuções incluídas nas médias

| Análise | Cenário | Início UTC (24/10/2026) | Publicadas / recebidas | Evidências |
|---|---|---|---:|---|
| A | MSG-EMQX-C1000-Q1-R1 | 00:46:24.4570334Z | 60000 / 13191 | [resumo](../results/MSG-EMQX-C1000-Q1-R1_emqx_20261024T004624456525000Z/summary.json) · [configuração](../results/MSG-EMQX-C1000-Q1-R1_emqx_20261024T004624456525000Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C1000-Q1-R1_emqx_20261024T004624456525000Z/) |
| A | MSG-EMQX-C1000-Q1-R2 | 00:56:17.5516185Z | 60000 / 13160 | [resumo](../results/MSG-EMQX-C1000-Q1-R2_emqx_20261024T005617550597600Z/summary.json) · [configuração](../results/MSG-EMQX-C1000-Q1-R2_emqx_20261024T005617550597600Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C1000-Q1-R2_emqx_20261024T005617550597600Z/) |
| A | MSG-EMQX-C1000-Q1-R3 | 00:58:06.4406285Z | 60000 / 13320 | [resumo](../results/MSG-EMQX-C1000-Q1-R3_emqx_20261024T005806439606300Z/summary.json) · [configuração](../results/MSG-EMQX-C1000-Q1-R3_emqx_20261024T005806439606300Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C1000-Q1-R3_emqx_20261024T005806439606300Z/) |
| A | MSG-EMQX-C100-Q1-R1 | 00:15:06.8480994Z | 6000 / 6000 | [resumo](../results/MSG-EMQX-C100-Q1-R1_emqx_20261024T001506847579000Z/summary.json) · [configuração](../results/MSG-EMQX-C100-Q1-R1_emqx_20261024T001506847579000Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C100-Q1-R1_emqx_20261024T001506847579000Z/) |
| A | MSG-EMQX-C100-Q1-R2 | 00:22:39.1446814Z | 6000 / 6000 | [resumo](../results/MSG-EMQX-C100-Q1-R2_emqx_20261024T002239143654800Z/summary.json) · [configuração](../results/MSG-EMQX-C100-Q1-R2_emqx_20261024T002239143654800Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C100-Q1-R2_emqx_20261024T002239143654800Z/) |
| A | MSG-EMQX-C100-Q1-R3 | 00:24:23.1846946Z | 6000 / 6000 | [resumo](../results/MSG-EMQX-C100-Q1-R3_emqx_20261024T002423184173200Z/summary.json) · [configuração](../results/MSG-EMQX-C100-Q1-R3_emqx_20261024T002423184173200Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C100-Q1-R3_emqx_20261024T002423184173200Z/) |
| B | MSG-EMQX-C500-Q0-R1 | 02:39:03.5212045Z | 30000 / 30000 | [resumo](../results/MSG-EMQX-C500-Q0-R1_emqx_20261024T023903520183200Z/summary.json) · [configuração](../results/MSG-EMQX-C500-Q0-R1_emqx_20261024T023903520183200Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C500-Q0-R1_emqx_20261024T023903520183200Z/) |
| B | MSG-EMQX-C500-Q0-R2 | 02:45:20.6796799Z | 29940 / 29940 | [resumo](../results/MSG-EMQX-C500-Q0-R2_emqx_20261024T024520677387100Z/summary.json) · [configuração](../results/MSG-EMQX-C500-Q0-R2_emqx_20261024T024520677387100Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C500-Q0-R2_emqx_20261024T024520677387100Z/) |
| B | MSG-EMQX-C500-Q0-R3 | 02:55:52.7395327Z | 30000 / 30000 | [resumo](../results/MSG-EMQX-C500-Q0-R3_emqx_20261024T025552738508300Z/summary.json) · [configuração](../results/MSG-EMQX-C500-Q0-R3_emqx_20261024T025552738508300Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C500-Q0-R3_emqx_20261024T025552738508300Z/) |
| A e B | MSG-EMQX-C500-Q1-R1 | 00:28:55.6727358Z | 30000 / 12951 | [resumo](../results/MSG-EMQX-C500-Q1-R1_emqx_20261024T002855672175100Z/summary.json) · [configuração](../results/MSG-EMQX-C500-Q1-R1_emqx_20261024T002855672175100Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C500-Q1-R1_emqx_20261024T002855672175100Z/) |
| A e B | MSG-EMQX-C500-Q1-R2 | 00:39:18.1419806Z | 30000 / 12911 | [resumo](../results/MSG-EMQX-C500-Q1-R2_emqx_20261024T003918140938000Z/summary.json) · [configuração](../results/MSG-EMQX-C500-Q1-R2_emqx_20261024T003918140938000Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C500-Q1-R2_emqx_20261024T003918140938000Z/) |
| A e B | MSG-EMQX-C500-Q1-R3 | 00:41:13.1362617Z | 30000 / 13104 | [resumo](../results/MSG-EMQX-C500-Q1-R3_emqx_20261024T004113135752700Z/summary.json) · [configuração](../results/MSG-EMQX-C500-Q1-R3_emqx_20261024T004113135752700Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C500-Q1-R3_emqx_20261024T004113135752700Z/) |
| B | MSG-EMQX-C500-Q2-R1 | 03:02:30.6495639Z | 30000 / 6538 | [resumo](../results/MSG-EMQX-C500-Q2-R1_emqx_20261024T030230648532300Z/summary.json) · [configuração](../results/MSG-EMQX-C500-Q2-R1_emqx_20261024T030230648532300Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C500-Q2-R1_emqx_20261024T030230648532300Z/) |
| B | MSG-EMQX-C500-Q2-R2 | 03:17:11.2814074Z | 30000 / 6560 | [resumo](../results/MSG-EMQX-C500-Q2-R2_emqx_20261024T031711280378900Z/summary.json) · [configuração](../results/MSG-EMQX-C500-Q2-R2_emqx_20261024T031711280378900Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C500-Q2-R2_emqx_20261024T031711280378900Z/) |
| B | MSG-EMQX-C500-Q2-R3 | 03:19:57.5912489Z | 30000 / 6538 | [resumo](../results/MSG-EMQX-C500-Q2-R3_emqx_20261024T031957589708800Z/summary.json) · [configuração](../results/MSG-EMQX-C500-Q2-R3_emqx_20261024T031957589708800Z/effective-config.conf) · [pasta](../results/MSG-EMQX-C500-Q2-R3_emqx_20261024T031957589708800Z/) |
| A | MSG-MOSQ-C1000-Q1-R1 | 00:48:52.8906175Z | 60000 / 7828 | [resumo](../results/MSG-MOSQ-C1000-Q1-R1_mosquitto_20261024T004852889511800Z/summary.json) · [configuração](../results/MSG-MOSQ-C1000-Q1-R1_mosquitto_20261024T004852889511800Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C1000-Q1-R1_mosquitto_20261024T004852889511800Z/) |
| A | MSG-MOSQ-C1000-Q1-R2 | 00:50:52.6615942Z | 59940 / 7820 | [resumo](../results/MSG-MOSQ-C1000-Q1-R2_mosquitto_20261024T005052661071400Z/summary.json) · [configuração](../results/MSG-MOSQ-C1000-Q1-R2_mosquitto_20261024T005052661071400Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C1000-Q1-R2_mosquitto_20261024T005052661071400Z/) |
| A | MSG-MOSQ-C1000-Q1-R3 | 01:00:44.7127433Z | 60000 / 7839 | [resumo](../results/MSG-MOSQ-C1000-Q1-R3_mosquitto_20261024T010044712233900Z/summary.json) · [configuração](../results/MSG-MOSQ-C1000-Q1-R3_mosquitto_20261024T010044712233900Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C1000-Q1-R3_mosquitto_20261024T010044712233900Z/) |
| A | MSG-MOSQ-C100-Q1-R1 | 00:16:55.1711226Z | 6000 / 6000 | [resumo](../results/MSG-MOSQ-C100-Q1-R1_mosquitto_20261024T001655170075000Z/summary.json) · [configuração](../results/MSG-MOSQ-C100-Q1-R1_mosquitto_20261024T001655170075000Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C100-Q1-R1_mosquitto_20261024T001655170075000Z/) |
| A | MSG-MOSQ-C100-Q1-R2 | 00:20:26.7420343Z | 6000 / 6000 | [resumo](../results/MSG-MOSQ-C100-Q1-R2_mosquitto_20261024T002026741524600Z/summary.json) · [configuração](../results/MSG-MOSQ-C100-Q1-R2_mosquitto_20261024T002026741524600Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C100-Q1-R2_mosquitto_20261024T002026741524600Z/) |
| A | MSG-MOSQ-C100-Q1-R3 | 00:26:33.0751033Z | 6000 / 6000 | [resumo](../results/MSG-MOSQ-C100-Q1-R3_mosquitto_20261024T002633074565200Z/summary.json) · [configuração](../results/MSG-MOSQ-C100-Q1-R3_mosquitto_20261024T002633074565200Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C100-Q1-R3_mosquitto_20261024T002633074565200Z/) |
| B | MSG-MOSQ-C500-Q0-R1 | 02:41:11.7761297Z | 30000 / 30000 | [resumo](../results/MSG-MOSQ-C500-Q0-R1_mosquitto_20261024T024111775087300Z/summary.json) · [configuração](../results/MSG-MOSQ-C500-Q0-R1_mosquitto_20261024T024111775087300Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C500-Q0-R1_mosquitto_20261024T024111775087300Z/) |
| B | MSG-MOSQ-C500-Q0-R2 | 02:43:14.719449Z | 30000 / 30000 | [resumo](../results/MSG-MOSQ-C500-Q0-R2_mosquitto_20261024T024314718386600Z/summary.json) · [configuração](../results/MSG-MOSQ-C500-Q0-R2_mosquitto_20261024T024314718386600Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C500-Q0-R2_mosquitto_20261024T024314718386600Z/) |
| B | MSG-MOSQ-C500-Q0-R3 | 02:57:53.455566Z | 30000 / 30000 | [resumo](../results/MSG-MOSQ-C500-Q0-R3_mosquitto_20261024T025753454509800Z/summary.json) · [configuração](../results/MSG-MOSQ-C500-Q0-R3_mosquitto_20261024T025753454509800Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C500-Q0-R3_mosquitto_20261024T025753454509800Z/) |
| A e B | MSG-MOSQ-C500-Q1-R1 | 00:31:02.4301353Z | 30000 / 7720 | [resumo](../results/MSG-MOSQ-C500-Q1-R1_mosquitto_20261024T003102429605600Z/summary.json) · [configuração](../results/MSG-MOSQ-C500-Q1-R1_mosquitto_20261024T003102429605600Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C500-Q1-R1_mosquitto_20261024T003102429605600Z/) |
| A e B | MSG-MOSQ-C500-Q1-R2 | 00:35:08.9599054Z | 30000 / 7800 | [resumo](../results/MSG-MOSQ-C500-Q1-R2_mosquitto_20261024T003508958868300Z/summary.json) · [configuração](../results/MSG-MOSQ-C500-Q1-R2_mosquitto_20261024T003508958868300Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C500-Q1-R2_mosquitto_20261024T003508958868300Z/) |
| A e B | MSG-MOSQ-C500-Q1-R3 | 00:43:54.3219991Z | 29940 / 7724 | [resumo](../results/MSG-MOSQ-C500-Q1-R3_mosquitto_20261024T004354320915300Z/summary.json) · [configuração](../results/MSG-MOSQ-C500-Q1-R3_mosquitto_20261024T004354320915300Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C500-Q1-R3_mosquitto_20261024T004354320915300Z/) |
| B | MSG-MOSQ-C500-Q2-R1 | 03:09:25.1532629Z | 30000 / 3400 | [resumo](../results/MSG-MOSQ-C500-Q2-R1_mosquitto_20261024T030925152728500Z/summary.json) · [configuração](../results/MSG-MOSQ-C500-Q2-R1_mosquitto_20261024T030925152728500Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C500-Q2-R1_mosquitto_20261024T030925152728500Z/) |
| B | MSG-MOSQ-C500-Q2-R2 | 03:13:50.2186245Z | 30000 / 3310 | [resumo](../results/MSG-MOSQ-C500-Q2-R2_mosquitto_20261024T031350218116100Z/summary.json) · [configuração](../results/MSG-MOSQ-C500-Q2-R2_mosquitto_20261024T031350218116100Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C500-Q2-R2_mosquitto_20261024T031350218116100Z/) |
| B | MSG-MOSQ-C500-Q2-R3 | 03:21:56.522153Z | 30000 / 3340 | [resumo](../results/MSG-MOSQ-C500-Q2-R3_mosquitto_20261024T032156521129200Z/summary.json) · [configuração](../results/MSG-MOSQ-C500-Q2-R3_mosquitto_20261024T032156521129200Z/effective-config.conf) · [pasta](../results/MSG-MOSQ-C500-Q2-R3_mosquitto_20261024T032156521129200Z/) |

<a id="internet-conexoes"></a>

### Conexões: execuções incluídas nas médias

| Cenário | Início UTC (24/10/2026) | Conectadas / solicitadas | Evidências |
|---|---|---:|---|
| CONN-EMQX-C1000-R1 | 03:53:29.5124336Z | 1000 / 1000 | [resumo](../results/CONN-EMQX-C1000-R1_emqx_20261024T035329511902800Z/summary.json) · [configuração](../results/CONN-EMQX-C1000-R1_emqx_20261024T035329511902800Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C1000-R1_emqx_20261024T035329511902800Z/) · [tentativas](../results/CONN-EMQX-C1000-R1_emqx_20261024T035329511902800Z/connections.csv) |
| CONN-EMQX-C1000-R2 | 04:02:33.8847043Z | 1000 / 1000 | [resumo](../results/CONN-EMQX-C1000-R2_emqx_20261024T040233884194100Z/summary.json) · [configuração](../results/CONN-EMQX-C1000-R2_emqx_20261024T040233884194100Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C1000-R2_emqx_20261024T040233884194100Z/) · [tentativas](../results/CONN-EMQX-C1000-R2_emqx_20261024T040233884194100Z/connections.csv) |
| CONN-EMQX-C1000-R3 | 04:13:54.4685031Z | 1000 / 1000 | [resumo](../results/CONN-EMQX-C1000-R3_emqx_20261024T041354467483300Z/summary.json) · [configuração](../results/CONN-EMQX-C1000-R3_emqx_20261024T041354467483300Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C1000-R3_emqx_20261024T041354467483300Z/) · [tentativas](../results/CONN-EMQX-C1000-R3_emqx_20261024T041354467483300Z/connections.csv) |
| CONN-EMQX-C2500-R1 | 06:43:52.5391523Z | 2499 / 2500 | [resumo](../results/CONN-EMQX-C2500-R1_emqx_20261024T064352538116300Z/summary.json) · [configuração](../results/CONN-EMQX-C2500-R1_emqx_20261024T064352538116300Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C2500-R1_emqx_20261024T064352538116300Z/) · [tentativas](../results/CONN-EMQX-C2500-R1_emqx_20261024T064352538116300Z/connections.csv) |
| CONN-EMQX-C2500-R2 | 06:50:27.9528767Z | 2499 / 2500 | [resumo](../results/CONN-EMQX-C2500-R2_emqx_20261024T065027951347400Z/summary.json) · [configuração](../results/CONN-EMQX-C2500-R2_emqx_20261024T065027951347400Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C2500-R2_emqx_20261024T065027951347400Z/) · [tentativas](../results/CONN-EMQX-C2500-R2_emqx_20261024T065027951347400Z/connections.csv) |
| CONN-EMQX-C2500-R3 | 07:03:40.1607093Z | 2500 / 2500 | [resumo](../results/CONN-EMQX-C2500-R3_emqx_20261024T070340159690900Z/summary.json) · [configuração](../results/CONN-EMQX-C2500-R3_emqx_20261024T070340159690900Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C2500-R3_emqx_20261024T070340159690900Z/) · [tentativas](../results/CONN-EMQX-C2500-R3_emqx_20261024T070340159690900Z/connections.csv) |
| CONN-EMQX-C500-R1 | 03:32:24.8125904Z | 500 / 500 | [resumo](../results/CONN-EMQX-C500-R1_emqx_20261024T033224811515100Z/summary.json) · [configuração](../results/CONN-EMQX-C500-R1_emqx_20261024T033224811515100Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C500-R1_emqx_20261024T033224811515100Z/) · [tentativas](../results/CONN-EMQX-C500-R1_emqx_20261024T033224811515100Z/connections.csv) |
| CONN-EMQX-C500-R2 | 03:39:58.3300884Z | 500 / 500 | [resumo](../results/CONN-EMQX-C500-R2_emqx_20261024T033958326938500Z/summary.json) · [configuração](../results/CONN-EMQX-C500-R2_emqx_20261024T033958326938500Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C500-R2_emqx_20261024T033958326938500Z/) · [tentativas](../results/CONN-EMQX-C500-R2_emqx_20261024T033958326938500Z/connections.csv) |
| CONN-EMQX-C500-R3 | 03:42:42.1830402Z | 499 / 500 | [resumo](../results/CONN-EMQX-C500-R3_emqx_20261024T034242182013200Z/summary.json) · [configuração](../results/CONN-EMQX-C500-R3_emqx_20261024T034242182013200Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C500-R3_emqx_20261024T034242182013200Z/) · [tentativas](../results/CONN-EMQX-C500-R3_emqx_20261024T034242182013200Z/connections.csv) |
| CONN-MOSQ-C1000-R1 | 03:55:47.0800696Z | 1000 / 1000 | [resumo](../results/CONN-MOSQ-C1000-R1_mosquitto_20261024T035547079028900Z/summary.json) · [configuração](../results/CONN-MOSQ-C1000-R1_mosquitto_20261024T035547079028900Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C1000-R1_mosquitto_20261024T035547079028900Z/) · [tentativas](../results/CONN-MOSQ-C1000-R1_mosquitto_20261024T035547079028900Z/connections.csv) |
| CONN-MOSQ-C1000-R2 | 04:00:02.6246056Z | 999 / 1000 | [resumo](../results/CONN-MOSQ-C1000-R2_mosquitto_20261024T040002624091300Z/summary.json) · [configuração](../results/CONN-MOSQ-C1000-R2_mosquitto_20261024T040002624091300Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C1000-R2_mosquitto_20261024T040002624091300Z/) · [tentativas](../results/CONN-MOSQ-C1000-R2_mosquitto_20261024T040002624091300Z/connections.csv) |
| CONN-MOSQ-C1000-R3 | 04:16:12.9567487Z | 1000 / 1000 | [resumo](../results/CONN-MOSQ-C1000-R3_mosquitto_20261024T041612955717700Z/summary.json) · [configuração](../results/CONN-MOSQ-C1000-R3_mosquitto_20261024T041612955717700Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C1000-R3_mosquitto_20261024T041612955717700Z/) · [tentativas](../results/CONN-MOSQ-C1000-R3_mosquitto_20261024T041612955717700Z/connections.csv) |
| CONN-MOSQ-C2500-R1 | 06:48:09.138911Z | 2500 / 2500 | [resumo](../results/CONN-MOSQ-C2500-R1_mosquitto_20261024T064809138374300Z/summary.json) · [configuração](../results/CONN-MOSQ-C2500-R1_mosquitto_20261024T064809138374300Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C2500-R1_mosquitto_20261024T064809138374300Z/) · [tentativas](../results/CONN-MOSQ-C2500-R1_mosquitto_20261024T064809138374300Z/connections.csv) |
| CONN-MOSQ-C2500-R2 | 07:00:19.8354024Z | 2499 / 2500 | [resumo](../results/CONN-MOSQ-C2500-R2_mosquitto_20261024T070019834384000Z/summary.json) · [configuração](../results/CONN-MOSQ-C2500-R2_mosquitto_20261024T070019834384000Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C2500-R2_mosquitto_20261024T070019834384000Z/) · [tentativas](../results/CONN-MOSQ-C2500-R2_mosquitto_20261024T070019834384000Z/connections.csv) |
| CONN-MOSQ-C2500-R3 | 07:07:07.9223061Z | 2499 / 2500 | [resumo](../results/CONN-MOSQ-C2500-R3_mosquitto_20261024T070707921780000Z/summary.json) · [configuração](../results/CONN-MOSQ-C2500-R3_mosquitto_20261024T070707921780000Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C2500-R3_mosquitto_20261024T070707921780000Z/) · [tentativas](../results/CONN-MOSQ-C2500-R3_mosquitto_20261024T070707921780000Z/connections.csv) |
| CONN-MOSQ-C500-R1 | 03:36:10.7822124Z | 500 / 500 | [resumo](../results/CONN-MOSQ-C500-R1_mosquitto_20261024T033610781702800Z/summary.json) · [configuração](../results/CONN-MOSQ-C500-R1_mosquitto_20261024T033610781702800Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C500-R1_mosquitto_20261024T033610781702800Z/) · [tentativas](../results/CONN-MOSQ-C500-R1_mosquitto_20261024T033610781702800Z/connections.csv) |
| CONN-MOSQ-C500-R2 | 03:38:04.667233Z | 500 / 500 | [resumo](../results/CONN-MOSQ-C500-R2_mosquitto_20261024T033804666128200Z/summary.json) · [configuração](../results/CONN-MOSQ-C500-R2_mosquitto_20261024T033804666128200Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C500-R2_mosquitto_20261024T033804666128200Z/) · [tentativas](../results/CONN-MOSQ-C500-R2_mosquitto_20261024T033804666128200Z/connections.csv) |
| CONN-MOSQ-C500-R3 | 03:49:09.0581147Z | 500 / 500 | [resumo](../results/CONN-MOSQ-C500-R3_mosquitto_20261024T034909057600000Z/summary.json) · [configuração](../results/CONN-MOSQ-C500-R3_mosquitto_20261024T034909057600000Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C500-R3_mosquitto_20261024T034909057600000Z/) · [tentativas](../results/CONN-MOSQ-C500-R3_mosquitto_20261024T034909057600000Z/connections.csv) |

### C2500: correção de descritores e exclusões

O processo Mosquitto operava com limite soft de 1.024 descritores.
O limite foi corrigido para 1.048.576 e os seis ensaios C2500 foram
repetidos, incluindo ambos os brokers. O procedimento está detalhado na
seção 12 dos [resultados do benchmark](relatorios/resultados-benchmark.md).
As execuções corrigidas são as de **06:43–07:08 UTC de 24/10/2026** da tabela
oficial. A coleta foi ampliada de 90 para **120 s**, para cobrir o hold
completo. Os arquivos `metadata.txt` dessas coletas não estão versionados.

Por exemplo, `CONN-MOSQ-C2500-R1` identifica tanto a tentativa de
04:21:42 UTC, com **1.017** handshakes, quanto a de 06:48:09 UTC, com
**2.500**. Somente a segunda entra nas médias. A mesma duplicação ocorre
em R2/R3 e nos três identificadores do EMQX.

| Cenário | Início UTC (24/10/2026) | Conectadas / solicitadas | Evidências |
|---|---|---:|---|
| CONN-EMQX-C2500-R1 | 04:18:50.2451295Z | 2499 / 2500 | [resumo](../results/CONN-EMQX-C2500-R1_emqx_20261024T041850244110000Z/summary.json) · [configuração](../results/CONN-EMQX-C2500-R1_emqx_20261024T041850244110000Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C2500-R1_emqx_20261024T041850244110000Z/) · [tentativas](../results/CONN-EMQX-C2500-R1_emqx_20261024T041850244110000Z/connections.csv) |
| CONN-EMQX-C2500-R2 | 04:27:51.0337172Z | 2499 / 2500 | [resumo](../results/CONN-EMQX-C2500-R2_emqx_20261024T042751033148000Z/summary.json) · [configuração](../results/CONN-EMQX-C2500-R2_emqx_20261024T042751033148000Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C2500-R2_emqx_20261024T042751033148000Z/) · [tentativas](../results/CONN-EMQX-C2500-R2_emqx_20261024T042751033148000Z/connections.csv) |
| CONN-EMQX-C2500-R3 | 04:30:03.2428021Z | 2499 / 2500 | [resumo](../results/CONN-EMQX-C2500-R3_emqx_20261024T043003241747100Z/summary.json) · [configuração](../results/CONN-EMQX-C2500-R3_emqx_20261024T043003241747100Z/effective-config.conf) · [pasta](../results/CONN-EMQX-C2500-R3_emqx_20261024T043003241747100Z/) · [tentativas](../results/CONN-EMQX-C2500-R3_emqx_20261024T043003241747100Z/connections.csv) |
| CONN-MOSQ-C2500-R1 | 04:21:42.5178491Z | 1017 / 2500 | [resumo](../results/CONN-MOSQ-C2500-R1_mosquitto_20261024T042142517322300Z/summary.json) · [configuração](../results/CONN-MOSQ-C2500-R1_mosquitto_20261024T042142517322300Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C2500-R1_mosquitto_20261024T042142517322300Z/) · [tentativas](../results/CONN-MOSQ-C2500-R1_mosquitto_20261024T042142517322300Z/connections.csv) |
| CONN-MOSQ-C2500-R2 | 04:25:55.6182302Z | 1017 / 2500 | [resumo](../results/CONN-MOSQ-C2500-R2_mosquitto_20261024T042555617718900Z/summary.json) · [configuração](../results/CONN-MOSQ-C2500-R2_mosquitto_20261024T042555617718900Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C2500-R2_mosquitto_20261024T042555617718900Z/) · [tentativas](../results/CONN-MOSQ-C2500-R2_mosquitto_20261024T042555617718900Z/connections.csv) |
| CONN-MOSQ-C2500-R3 | 04:35:06.2727218Z | 1017 / 2500 | [resumo](../results/CONN-MOSQ-C2500-R3_mosquitto_20261024T043506272213400Z/summary.json) · [configuração](../results/CONN-MOSQ-C2500-R3_mosquitto_20261024T043506272213400Z/effective-config.conf) · [pasta](../results/CONN-MOSQ-C2500-R3_mosquitto_20261024T043506272213400Z/) · [tentativas](../results/CONN-MOSQ-C2500-R3_mosquitto_20261024T043506272213400Z/connections.csv) |

### Diagnósticos QoS: fora das médias

Esses quatro testes mantêm C500, 60 s de publicação e ampliam o drain para
180 s (Q1) ou 420 s (Q2). Não são repetições adicionais dos experimentos A/B.

| Cenário | Início UTC (24/10/2026) | Drain | Publicadas / recebidas | Evidências |
|---|---|---:|---:|---|
| DIAG-EMQX-C500-Q1 | 07:45:54.7579215Z | 180 s | 30000 / 13000 | [resumo](../results/DIAG-EMQX-C500-Q1_emqx_20261024T074554757518600Z/summary.json) · [configuração](../results/DIAG-EMQX-C500-Q1_emqx_20261024T074554757518600Z/effective-config.conf) · [pasta](../results/DIAG-EMQX-C500-Q1_emqx_20261024T074554757518600Z/) |
| DIAG-EMQX-C500-Q2 | 07:50:55.9879077Z | 420 s | 30000 / 7016 | [resumo](../results/DIAG-EMQX-C500-Q2_emqx_20261024T075055986879500Z/summary.json) · [configuração](../results/DIAG-EMQX-C500-Q2_emqx_20261024T075055986879500Z/effective-config.conf) · [pasta](../results/DIAG-EMQX-C500-Q2_emqx_20261024T075055986879500Z/) |
| DIAG-MOSQ-C500-Q1 | 08:00:08.3856897Z | 180 s | 30000 / 8144 | [resumo](../results/DIAG-MOSQ-C500-Q1_mosquitto_20261024T080008381260600Z/summary.json) · [configuração](../results/DIAG-MOSQ-C500-Q1_mosquitto_20261024T080008381260600Z/effective-config.conf) · [pasta](../results/DIAG-MOSQ-C500-Q1_mosquitto_20261024T080008381260600Z/) |
| DIAG-MOSQ-C500-Q2 | 08:04:31.285679Z | 420 s | 30000 / 4021 | [resumo](../results/DIAG-MOSQ-C500-Q2_mosquitto_20261024T080431285149300Z/summary.json) · [configuração](../results/DIAG-MOSQ-C500-Q2_mosquitto_20261024T080431285149300Z/effective-config.conf) · [pasta](../results/DIAG-MOSQ-C500-Q2_mosquitto_20261024T080431285149300Z/) |

<a id="aws-mensagens"></a>

## Mensagens em rede privada AWS

São **seis execuções exploratórias**, uma por broker/carga, separadas da
matriz pela Internet. Parâmetros: QoS 1, 1 msg/s por publicador, 256 B,
`staggered`, abertura solicitada de 100 conexões/s, warmup 10 s, medição
60 s e **drain 60 s**, com um único consumidor.
O gerador tinha aproximadamente 2 GiB em C2500 e foi ampliado para 8 GiB
antes dos resultados C5000/C10000 aqui selecionados.

Os arquivos de [modelos de estresse](../configs/tcc-mqtt-stress-configs/messages/)
com `drain=5s` são **modelos**, não o registro dos seis resultados AWS.
O `effective-config.conf` vinculado em cada linha registra a configuração
da execução, com `drain=1m0s` e `drain_seconds=60` no `summary.json`.
As tentativas anteriores de C5000 não substituem essas execuções.

| Publicadores / broker | run_id / início UTC | Publicadas / recebidas | Resultado e configuração | Telemetria do broker / duração da coleta |
|---|---|---:|---|---|
| 2500 / emqx | `20261107T043233.547254580Z`<br>2026-11-07T04:32:33.547590957Z | 150000 / 150000 | [resumo](../results/STRESS-MSG-EMQX-C2500-Q1-AWS-R1_emqx_20261107T043233547254580Z/summary.json) · [configuração](../results/STRESS-MSG-EMQX-C2500-Q1-AWS-R1_emqx_20261107T043233547254580Z/effective-config.conf) · [pasta](../results/STRESS-MSG-EMQX-C2500-Q1-AWS-R1_emqx_20261107T043233547254580Z/) | [arquivos](../broker-metrics/STRESS-MSG-EMQX-C2500-Q1-AWS-R1_emqx_20261107T043233547254580Z/) · [metadados](../broker-metrics/STRESS-MSG-EMQX-C2500-Q1-AWS-R1_emqx_20261107T043233547254580Z/metadata.txt) · 220 s |
| 2500 / mosquitto | `20261107T043727.799683781Z`<br>2026-11-07T04:37:27.800043795Z | 149998 / 149998 | [resumo](../results/STRESS-MSG-MOSQ-C2500-Q1-AWS-R1_mosquitto_20261107T043727799683781Z/summary.json) · [configuração](../results/STRESS-MSG-MOSQ-C2500-Q1-AWS-R1_mosquitto_20261107T043727799683781Z/effective-config.conf) · [pasta](../results/STRESS-MSG-MOSQ-C2500-Q1-AWS-R1_mosquitto_20261107T043727799683781Z/) | [arquivos](../broker-metrics/STRESS-MSG-MOSQ-C2500-Q1-AWS-R1_mosquitto_20261107T043727799683781Z/) · [metadados](../broker-metrics/STRESS-MSG-MOSQ-C2500-Q1-AWS-R1_mosquitto_20261107T043727799683781Z/metadata.txt) · 210 s |
| 5000 / emqx | `20261107T134300.946027406Z`<br>2026-11-07T13:43:00.94766886Z | 299996 / 296117 | [resumo](../results/STRESS-MSG-EMQX-C5000-Q1-AWS-R1_emqx_20261107T134300946027406Z/summary.json) · [configuração](../results/STRESS-MSG-EMQX-C5000-Q1-AWS-R1_emqx_20261107T134300946027406Z/effective-config.conf) · [pasta](../results/STRESS-MSG-EMQX-C5000-Q1-AWS-R1_emqx_20261107T134300946027406Z/) | [arquivos](../broker-metrics/STRESS-MSG-EMQX-C5000-Q1-AWS-R1_emqx_20261107T134300946027406Z/) · [metadados](../broker-metrics/STRESS-MSG-EMQX-C5000-Q1-AWS-R1_emqx_20261107T134300946027406Z/metadata.txt) · 240 s |
| 5000 / mosquitto | `20261107T134735.989680758Z`<br>2026-11-07T13:47:35.98990671Z | 300000 / 300000 | [resumo](../results/STRESS-MSG-MOSQ-C5000-Q1-AWS-R1_mosquitto_20261107T134735989680758Z/summary.json) · [configuração](../results/STRESS-MSG-MOSQ-C5000-Q1-AWS-R1_mosquitto_20261107T134735989680758Z/effective-config.conf) · [pasta](../results/STRESS-MSG-MOSQ-C5000-Q1-AWS-R1_mosquitto_20261107T134735989680758Z/) | [arquivos](../broker-metrics/STRESS-MSG-MOSQ-C5000-Q1-AWS-R1_mosquitto_20261107T134735989680758Z/) · [metadados](../broker-metrics/STRESS-MSG-MOSQ-C5000-Q1-AWS-R1_mosquitto_20261107T134735989680758Z/metadata.txt) · 240 s |
| 10000 / emqx | `20261107T135159.299404144Z`<br>2026-11-07T13:51:59.299670254Z | 599999 / 0 | [resumo](../results/STRESS-MSG-EMQX-C10000-Q1-AWS-R1_emqx_20261107T135159299404144Z/summary.json) · [configuração](../results/STRESS-MSG-EMQX-C10000-Q1-AWS-R1_emqx_20261107T135159299404144Z/effective-config.conf) · [pasta](../results/STRESS-MSG-EMQX-C10000-Q1-AWS-R1_emqx_20261107T135159299404144Z/) | [arquivos](../broker-metrics/STRESS-MSG-EMQX-C10000-Q1-AWS-R1_emqx_20261107T135159299404144Z/) · [metadados](../broker-metrics/STRESS-MSG-EMQX-C10000-Q1-AWS-R1_emqx_20261107T135159299404144Z/metadata.txt) · 280 s |
| 10000 / mosquitto | `20261107T135759.594872583Z`<br>2026-11-07T13:57:59.595091819Z | 599998 / 564206 | [resumo](../results/STRESS-MSG-MOSQ-C10000-Q1-AWS-R1_mosquitto_20261107T135759594872583Z/summary.json) · [configuração](../results/STRESS-MSG-MOSQ-C10000-Q1-AWS-R1_mosquitto_20261107T135759594872583Z/effective-config.conf) · [pasta](../results/STRESS-MSG-MOSQ-C10000-Q1-AWS-R1_mosquitto_20261107T135759594872583Z/) | [arquivos](../broker-metrics/STRESS-MSG-MOSQ-C10000-Q1-AWS-R1_mosquitto_20261107T135759594872583Z/) · [metadados](../broker-metrics/STRESS-MSG-MOSQ-C10000-Q1-AWS-R1_mosquitto_20261107T135759594872583Z/metadata.txt) · 280 s |

Cada pasta de telemetria contém `cpu.txt`, `memory.txt`, `network.txt`,
`process.txt`, `connections.csv` e `metadata.txt`. O CSV de conexões do
broker é amostragem TCP; não é o registro individual de tentativas do
gerador. A duração da tabela vem de `duration_seconds` da **coleta**,
distinta dos 60 s de publicação. Use os horários do resumo e das amostras
para selecionar a fase analisada; a média do arquivo inteiro inclui
abertura, warmup, publicação, drain e períodos ociosos.

<a id="emqx-retestes"></a>

### EMQX C10000: execução original e dois retestes

Todos usaram `scenario_id=STRESS-MSG-EMQX-C10000-Q1-AWS-R1` e
`repetition=1`. São execuções distintas. Os ajustes abaixo são do
**broker**; não são parâmetros do arquivo `effective-config.conf` do gerador.

| Execução / `run_id` | Fila `mqtt.max_mqueue_len` | Mailbox `force_shutdown.max_mailbox_size` | `mqtt.max_inflight` | Publicações confirmadas / recebidas na medição | Evidência preservada |
|---|---:|---:|---:|---:|---|
| Original `20261107T135159.299404144Z` | 1.000 | 1.000 | 32 | 599.999 / 0 | Resumo e métricas na tabela AWS; evento de 13:53:41 UTC: fila cheia e `mailbox_overflow`, valor 1.484 |
| Reteste 1 `20261107T155303.390657397Z` | 10.000 | 10.000 | 32 | 599.999 / 0 | Fila cheia às 15:54:47 UTC e eventos de log suprimidos às 15:55:17 UTC |
| Reteste 2 `20261107T161242.481842647Z` | 10.000 | 10.000 | 256 | 599.999 / 0 | Fila cheia e eventos de log suprimidos entre 16:14:24 e 16:15:26 UTC |

As configurações, os indicadores e os eventos de log estão apresentados nas
seções 9–11 dos [resultados de mensagens](relatorios/resultados-mensagens.md).
No reteste 1, a proteção estava ativa e a leitura posterior mostrou
`max_heap_size=32MB`. No reteste 2 foram confirmados
`force_shutdown.enable=true` e `force_shutdown.max_heap_size="32MB"`.
Os arquivos individuais `summary.json`, os logs completos e as séries de
CPU/rede dos dois retestes não estão versionados. A análise desses ensaios
se limita aos indicadores e eventos apresentados na documentação.

Nos dois retestes há registro de fila cheia. O segundo não apresenta novo
`mailbox_overflow` nos trechos disponíveis, nem prova de que o consumidor
permaneceu conectado até o final. Eventos de log suprimidos por throttling
não equivalem automaticamente ao número de mensagens perdidas.
Zero mensagens contabilizadas na fase `measure` não prova descarte de
todas as publicações. A proteção interna `check_oom_shutdown` do EMQX
nesses logs é distinta do OOM Killer do Linux no ensaio 50k.

<a id="aws-50k"></a>

## 50.000 conexões em rede privada AWS

Quatro execuções exploratórias após a ampliação de portas efêmeras do
gerador. R4 usa 1.000 tentativas/s; R5 usa taxa 0 (sem limitador); hold
configurado de 60 s. Em todos os resumos, `repetition=1`, apesar dos sufixos
R4/R5. A inconsistência foi preservada: não são quatro/cinco repetições
homogêneas nem entram na matriz oficial.

| Cenário / início UTC | Taxa solicitada | Handshakes / falhas | Resultado e configuração | Broker / duração da coleta | Gerador |
|---|---:|---:|---|---|---|
| STRESS-CONN-EMQX-C50000-R4<br>2026-11-07T03:26:01.458016324Z | 1000/s | 50000 / 0 | [resumo](../results/STRESS-CONN-EMQX-C50000-R4_emqx_20261107T032601457253253Z/summary.json) · [configuração](../results/STRESS-CONN-EMQX-C50000-R4_emqx_20261107T032601457253253Z/effective-config.conf) · [pasta](../results/STRESS-CONN-EMQX-C50000-R4_emqx_20261107T032601457253253Z/) · [tentativas](../results/STRESS-CONN-EMQX-C50000-R4_emqx_20261107T032601457253253Z/connections.csv) | [arquivos](../broker-metrics/STRESS-CONN-EMQX-C50000-R4_emqx_20261107T032601457253253Z/) · [metadados](../broker-metrics/STRESS-CONN-EMQX-C50000-R4_emqx_20261107T032601457253253Z/metadata.txt) · 200 s | [arquivos](../generator-metrics/STRESS-CONN-EMQX-C50000-R4_emqx_20261107T032601457253253Z/) |
| STRESS-CONN-MOSQ-C50000-R4<br>2026-11-07T03:38:35.740291476Z | 1000/s | 50000 / 0 | [resumo](../results/STRESS-CONN-MOSQ-C50000-R4_mosquitto_20261107T033835740064344Z/summary.json) · [configuração](../results/STRESS-CONN-MOSQ-C50000-R4_mosquitto_20261107T033835740064344Z/effective-config.conf) · [pasta](../results/STRESS-CONN-MOSQ-C50000-R4_mosquitto_20261107T033835740064344Z/) · [tentativas](../results/STRESS-CONN-MOSQ-C50000-R4_mosquitto_20261107T033835740064344Z/connections.csv) | [arquivos](../broker-metrics/STRESS-CONN-MOSQ-C50000-R4_mosquitto_20261107T033835740064344Z/) · [metadados](../broker-metrics/STRESS-CONN-MOSQ-C50000-R4_mosquitto_20261107T033835740064344Z/metadata.txt) · 200 s | [arquivos](../generator-metrics/STRESS-CONN-MOSQ-C50000-R4_mosquitto_20261107T033835740064344Z/) |
| STRESS-CONN-EMQX-C50000-R5<br>2026-11-07T03:44:08.538022944Z | 0/s | 31263 / 18737 | [resumo](../results/STRESS-CONN-EMQX-C50000-R5_emqx_20261107T034408537793201Z/summary.json) · [configuração](../results/STRESS-CONN-EMQX-C50000-R5_emqx_20261107T034408537793201Z/effective-config.conf) · [pasta](../results/STRESS-CONN-EMQX-C50000-R5_emqx_20261107T034408537793201Z/) · [tentativas](../results/STRESS-CONN-EMQX-C50000-R5_emqx_20261107T034408537793201Z/connections.csv) | [arquivos](../broker-metrics/STRESS-CONN-EMQX-C50000-R5_emqx_20261107T034408537793201Z/) · [metadados](../broker-metrics/STRESS-CONN-EMQX-C50000-R5_emqx_20261107T034408537793201Z/metadata.txt) · 200 s | [arquivos](../generator-metrics/STRESS-CONN-EMQX-C50000-R5_emqx_20261107T034408537793201Z/) |
| STRESS-CONN-MOSQ-C50000-R5<br>2026-11-07T03:51:12.675191018Z | 0/s | 31784 / 18216 | [resumo](../results/STRESS-CONN-MOSQ-C50000-R5_mosquitto_20261107T035112674907938Z/summary.json) · [configuração](../results/STRESS-CONN-MOSQ-C50000-R5_mosquitto_20261107T035112674907938Z/effective-config.conf) · [pasta](../results/STRESS-CONN-MOSQ-C50000-R5_mosquitto_20261107T035112674907938Z/) · [tentativas](../results/STRESS-CONN-MOSQ-C50000-R5_mosquitto_20261107T035112674907938Z/connections.csv) | [arquivos](../broker-metrics/STRESS-CONN-MOSQ-C50000-R5_mosquitto_20261107T035112674907938Z/) · [metadados](../broker-metrics/STRESS-CONN-MOSQ-C50000-R5_mosquitto_20261107T035112674907938Z/metadata.txt) · 200 s | [arquivos](../generator-metrics/STRESS-CONN-MOSQ-C50000-R5_mosquitto_20261107T035112674907938Z/) |

As pastas do gerador contêm `sistema.txt`, `processo.txt`, `conexoes.csv`
e `oom.txt`. Os quatro `oom.txt` estão vazios: não houve registro de OOM
nesses arquivos da EC2 geradora. O OOM do servidor EMQX é um evento distinto.

Os [resultados de 50.000 conexões, seção 2](relatorios/resultados-conexoes-50k.md)
apresentam os eventos do kernel e systemd que registram OOM no **broker
EMQX** às 03:45:00 UTC, em R5. O arquivo completo de `journalctl` não está
versionado. Os 31.263 handshakes
contabilizados em R5 não atestam conexões sustentadas durante todo o hold.
O gerador chegou ao limite de CPU nas quatro condições; as taxas efetivas
não medem isoladamente a capacidade máxima dos brokers. Em R4 EMQX,
o período de monitoramento do gerador posterior ao teste foi excluído da análise.

## Organização dos arquivos e limites de reprodução

Os arquivos dos experimentos estão organizados da seguinte forma:

- [`results/`](../results/): resumos JSON/CSV, configurações efetivas e tentativas individuais de conexão.
- [`broker-metrics/`](../broker-metrics/): séries de CPU, memória, rede e conexões TCP dos brokers nos ensaios AWS selecionados.
- [`generator-metrics/`](../generator-metrics/): séries de CPU, memória e conexões TCP do gerador nos quatro ensaios de 50.000 conexões.

O [manifesto de integridade](integridade-artefatos.csv) relaciona os arquivos
de dados dos dez ensaios AWS selecionados, seus tamanhos e hashes SHA-256.
Os metadados dos resultados permanecem como registrados em cada execução;
as diferenças de identificação e os critérios de inclusão estão descritos
nas respectivas seções.

Os resumos permitem conferir as contagens, as taxas e os indicadores de
latência consolidados. `messages.csv` e `messages.csv.gz` não estão
versionados por tamanho; a reprodução das distribuições por mensagem
exige esses registros individuais. Para os seis ensaios AWS de mensagens,
estão versionadas as séries de recursos dos brokers, sem um perfil completo
de CPU/RAM do gerador.

Os demais diretórios de `results/` incluem pilotos, estresse Windows e
outras cargas AWS, fora das seleções deste índice. Dois diretórios contêm
apenas `effective-config.conf`, sem resumo de resultado concluído:

- `results/STRESS-CONN-EMQX-C5000-R2_emqx_20261106T023848786655800Z/`;
- `results/STRESS-CONN-MOSQ-C15000-R1_mosquitto_20261106T185321060733372Z/`.

Esses dois diretórios não foram usados nas médias nem nas comparações
apresentadas.
