# Resultados do benchmark MQTT pela Internet

## 1. Finalidade deste documento

Este documento apresenta o método, os resultados e as limitações da matriz experimental pela Internet do TCC **“Projeto e avaliação de desempenho de uma plataforma IoT baseada em MQTT com EMQX”**. As execuções de cada cenário estão identificadas no [índice de experimentos](../experimentos-tcc.md).

A comparação experimental foi realizada entre **EMQX** e **Eclipse Mosquitto**, com foco em três dimensões principais:

1. desempenho na troca de mensagens MQTT;
2. consumo de recursos computacionais;
3. volume de conexões simultâneas.

O backend, banco de dados e frontend foram mantidos **fora do caminho de benchmark**, de forma que as medições desta etapa representassem o comportamento da comunicação MQTT e dos brokers, sem interferência da persistência ou da interface da plataforma.

---

## 2. Ambiente experimental

### 2.1 Infraestrutura

Foram utilizadas duas instâncias equivalentes na AWS, uma para cada broker:

- tipo de instância: **EC2 t3.small**;
- sistema operacional: **Ubuntu 24.04 LTS**;
- 2 vCPUs;
- aproximadamente 2 GiB de RAM;
- armazenamento gp3;
- instalações nativas dos brokers, sem Docker;
- mesma base de sistema utilizada para reduzir diferenças de ambiente.

Brokers avaliados:

- **EMQX Community 5.8.9**;
- **Eclipse Mosquitto 2.1.2**.

O gerador de carga foi executado em um computador Windows externo às instâncias EC2, comunicando-se com os brokers pela Internet.

Consequentemente, a latência analisada nesta etapa é a latência MQTT fim a fim no ambiente experimental, incluindo os percursos de rede entre gerador e broker.

---

## 3. Ferramenta de geração de carga

Foi utilizada a ferramenta própria **mqttbench v0.3.0**, desenvolvida em Go e utilizando MQTT 5.0.

A ferramenta possui dois modos principais.

### 3.1 Modo `message`

Nesse modo são criados múltiplos clientes publicadores e um consumidor independente responsável por receber as mensagens e calcular a latência.

A mensagem contém o instante de envio, permitindo que o cálculo da latência seja realizado no mesmo computador que gera a carga e recebe as mensagens, reduzindo problemas de sincronização de relógio entre máquinas.

Principais informações produzidas:

- clientes solicitados e conectados;
- tentativas de publicação;
- publicações realizadas;
- mensagens únicas recebidas;
- mensagens não observadas pelo coletor;
- duplicatas;
- erros;
- throughput de publicação;
- throughput de recebimento;
- latência mínima;
- latência média;
- p50;
- p95;
- p99;
- latência máxima.

### 3.2 Modo `connections`

Nesse modo a ferramenta cria conexões MQTT progressivamente, em uma taxa configurável, mantém as conexões por um período determinado e registra:

- conexões solicitadas;
- conexões estabelecidas;
- falhas de conexão;
- taxa efetiva de conexões por segundo;
- tempo da janela de estabelecimento;
- latência média de conexão;
- p95;
- p99;
- latência máxima.

---

## 4. Coleta de recursos das instâncias

Durante os testes foi executado um coletor independente nas instâncias EC2.

As principais fontes utilizadas foram:

- `sar -u` para CPU;
- `sar -r` para memória;
- `sar -n DEV` para rede;
- `pidstat` para CPU e memória do processo do broker;
- `ss` para acompanhar conexões TCP estabelecidas na porta MQTT.

Foram consolidadas métricas como:

- CPU média da EC2;
- pico de CPU da EC2;
- CPU média do processo do broker;
- pico de CPU do broker;
- RAM média da EC2;
- RAM média do processo do broker;
- tráfego médio RX;
- tráfego médio TX;
- quantidade de conexões observadas.

---

## 5. Matriz experimental oficial

Foram realizadas **48 execuções oficiais**, divididas em três experimentos.

Cada cenário foi executado **três vezes**.

### 5.1 Experimento A — crescimento da carga

Objetivo: analisar o comportamento dos brokers com o aumento do número de clientes publicadores.

Parâmetros:

- QoS: 1;
- clientes: 100, 500 e 1000;
- taxa: 1 mensagem/s por cliente;
- payload: 256 bytes;
- padrão de publicação: `staggered`;
- warmup: 10 s;
- duração oficial: 60 s;
- drain original: 5 s.

Total:

- 3 níveis de carga;
- 2 brokers;
- 3 repetições;
- **18 execuções**.

### 5.2 Experimento B — impacto do QoS

Objetivo: analisar a influência dos níveis de QoS sobre o desempenho.

Parâmetros:

- clientes: 500;
- QoS: 0, 1 e 2;
- taxa: 1 mensagem/s por cliente;
- payload: 256 bytes;
- publicação `staggered`;
- warmup: 10 s;
- duração: 60 s;
- drain original: 5 s.

O cenário C500/Q1 já fazia parte do Experimento A e foi reutilizado na análise.

Foram adicionadas:

- QoS 0;
- QoS 2;
- 2 brokers;
- 3 repetições.

Total adicional:

- **12 execuções**.

Total de testes de mensagens dos experimentos A e B:

- **30 execuções**.

### 5.3 Experimento C — volume de conexões

Objetivo: avaliar o comportamento dos brokers com o crescimento do número de conexões MQTT simultâneas.

Parâmetros:

- 500, 1000 e 2500 conexões;
- taxa solicitada: 100 conexões/s;
- hold: 60 s;
- 2 brokers;
- 3 repetições.

Total:

- **18 execuções**.

### 5.4 Total

| Categoria | Execuções |
|---|---:|
| Mensagens | 30 |
| Conexões | 18 |
| **Total oficial** | **48** |

Além das 48 execuções oficiais, foram realizados **4 testes diagnósticos complementares de QoS**, que não fazem parte da matriz oficial.

---

## 6. Procedimento de execução

Antes de cada execução foram realizadas verificações para evitar interferência de conexões anteriores e identificar possíveis anomalias no ambiente.

Entre as verificações utilizadas estavam:

- ausência de conexões MQTT residuais;
- observação da CPU em repouso;
- teste de conectividade entre o gerador e o broker;
- execução do coletor de recursos antes do início do `mqttbench`.

A ordem dos brokers foi alternada entre repetições para reduzir o risco de viés causado por mudanças temporais de rede ou infraestrutura.

Exemplo de alternância:

- R1: EMQX → Mosquitto;
- R2: Mosquitto → EMQX;
- R3: EMQX → Mosquitto.

O critério de exclusão foi a existência de causa técnica verificável, como configuração incorreta, interrupção do teste, coleta incompleta ou limitação externa ao objeto avaliado. Desempenho baixo, isoladamente, não constituiu motivo de exclusão.

---

## 7. Resultados — crescimento da carga com QoS 1

### 7.1 Throughput de recebimento

Médias das três repetições:

| Clientes | EMQX (msg/s) | Mosquitto (msg/s) |
|---:|---:|---:|
| 100 | 100,00 | 100,00 |
| 500 | 216,48 | 129,13 |
| 1000 | 220,39 | 130,48 |

Com 100 clientes, ambos os brokers acompanharam integralmente a taxa produzida.

Nos cenários de 500 e 1000 clientes com QoS 1, a taxa de recebimento do consumidor passou a ficar limitada, apesar de a taxa de publicação solicitada continuar crescendo.

No ambiente adotado, o EMQX apresentou throughput de recebimento superior ao Mosquitto nesses cenários.

O resultado caracteriza o ambiente avaliado. Os mecanismos e valores de controle de fluxo de QoS também influenciam o comportamento observado.

### 7.2 Entrega observada

| Clientes | EMQX (%) | Mosquitto (%) |
|---:|---:|---:|
| 100 | 100,00 | 100,00 |
| 500 | 43,30 | 25,84 |
| 1000 | 22,04 | 13,05 |

A redução do percentual recebido com o aumento da carga motivou testes diagnósticos adicionais, descritos na Seção 10.

---

## 8. Resultados — impacto do QoS com 500 clientes

### 8.1 Throughput de recebimento

| QoS | EMQX (msg/s) | Mosquitto (msg/s) |
|---:|---:|---:|
| 0 | 499,67 | 500,00 |
| 1 | 216,48 | 129,13 |
| 2 | 109,09 | 55,83 |

O QoS 0 apresentou comportamento próximo à taxa total de 500 mensagens/s nos dois brokers.

A utilização de QoS 1 provocou redução significativa da taxa de recebimento, e o QoS 2 apresentou redução ainda maior.

Esse comportamento é compatível com o aumento do custo de confirmação e controle de fluxo associado aos níveis superiores de QoS.

### 8.2 Entrega observada

| QoS | EMQX (%) | Mosquitto (%) |
|---:|---:|---:|
| 0 | 100,00 | 100,00 |
| 1 | 43,30 | 25,84 |
| 2 | 21,82 | 11,17 |

A expressão **“entrega observada”** é utilizada propositalmente.

O campo `lost` representa a diferença entre mensagens publicadas e mensagens efetivamente observadas pelo coletor durante a execução. Esse indicador não identifica a causa da ausência de recebimento nem comprova perda definitiva no broker.

---

## 9. Consumo de recursos

Os resultados demonstraram diferenças relevantes entre os brokers quanto ao consumo de recursos.

### 9.1 QoS 1 — RAM do processo

| Clientes | EMQX (MB) | Mosquitto (MB) |
|---:|---:|---:|
| 100 | 259,50 | 9,58 |
| 500 | 283,24 | 12,20 |
| 1000 | 320,29 | 14,62 |

O Mosquitto apresentou um footprint de memória substancialmente menor no ambiente avaliado.

### 9.2 QoS 1 — CPU média do processo

| Clientes | EMQX (%) | Mosquitto (%) |
|---:|---:|---:|
| 100 | 5,34 | 1,21 |
| 500 | 13,93 | 2,67 |
| 1000 | 25,74 | 4,61 |

Também foi observado menor consumo médio de CPU do processo do Mosquitto.

Esses resultados mostram que o Mosquitto possui vantagem relevante em economia de recursos no ambiente experimental.

A avaliação da plataforma inclui, além do consumo de recursos, os requisitos de administração, segurança, gerenciamento, observabilidade e integração da solução proposta.

---

## 10. Diagnóstico dos cenários QoS 1 e QoS 2

A redução da entrega observada em C500/Q1, C500/Q2 e C1000/Q1 levantou inicialmente a hipótese de que as mensagens permaneciam apenas em backlog e que um período maior de drain permitiria sua entrega posterior.

Para verificar essa hipótese foram executados quatro testes diagnósticos fora da matriz oficial.

### 10.1 Configuração dos diagnósticos

Foram mantidos:

- 500 clientes;
- 1 mensagem/s por cliente;
- payload de 256 bytes;
- 60 s de publicação;
- publicação `staggered`.

Somente o período de drain foi ampliado:

- QoS 1: 180 s;
- QoS 2: 420 s.

### 10.2 Resultados diagnósticos

| Cenário | Recebidas | Publicadas | Entrega observada | Throughput recebido |
|---|---:|---:|---:|---:|
| EMQX C500 Q1 | 13.000 | 30.000 | 43,33% | 216,67 msg/s |
| Mosquitto C500 Q1 | 8.144 | 30.000 | 27,15% | 135,73 msg/s |
| EMQX C500 Q2 | 7.016 | 30.000 | 23,39% | 116,93 msg/s |
| Mosquitto C500 Q2 | 4.021 | 30.000 | 13,40% | 67,02 msg/s |

A ampliação do drain **não restaurou a entrega para 100%**.

Portanto, a explicação “as mensagens apenas precisam de mais tempo para sair da fila” não foi suficiente para explicar o comportamento.

### 10.3 Interpretação

Os resultados são compatíveis com limitações associadas ao controle de fluxo de QoS 1/2 e às filas de entrega para um único consumidor remoto.

Durante a análise técnica foram considerados, entre outros, os parâmetros padrão de mensagens inflight e filas dos brokers.

Como o consumidor do benchmark é uma única conexão remota, a capacidade de escoamento dessa conexão também faz parte do sistema medido.

O indicador `lost` representa as mensagens não observadas pelo coletor durante a execução. A taxa de entrega observada e o throughput recebido descrevem o escoamento no cenário adotado; esses indicadores não determinam o destino individual das mensagens não recebidas.

---

## 11. Resultados — volume de conexões

Após a correção metodológica descrita na Seção 12, os resultados consolidados foram:

| Conexões solicitadas | EMQX sucesso | Mosquitto sucesso |
|---:|---:|---:|
| 500 | 99,93% | 100,00% |
| 1000 | 100,00% | 99,97% |
| 2500 | 99,97% | 99,97% |

Os dois brokers conseguiram manter aproximadamente 2500 conexões simultâneas no ambiente experimental.

### 11.1 C2500 corrigido

Médias:

| Métrica | EMQX | Mosquitto |
|---|---:|---:|
| Conectadas | 2499,33 | 2499,33 |
| Sucesso | 99,973% | 99,973% |
| Taxa efetiva | 86,38 conn/s | 97,23 conn/s |
| Latência média de conexão | 311,87 ms | 298,97 ms |
| p95 | 315,03 ms | 303,71 ms |
| CPU EC2 no hold | 1,38% | 0,91% |
| CPU broker no hold | 2,61% | 0,06% |
| RAM EC2 no hold | 355,62 MB | 186,83 MB |
| RAM do processo | 290,74 MB | 21,20 MB |

No cenário C2500 corrigido, ambos apresentaram capacidade de conexão praticamente equivalente em termos de sucesso.

O Mosquitto apresentou menor consumo de recursos e, nessa execução específica, maior taxa efetiva média de estabelecimento de conexões.

Esses valores caracterizam o ambiente experimental adotado, com comunicação pela Internet entre o gerador e as instâncias.

---

## 12. Correção metodológica — RLIMIT_NOFILE do Mosquitto

Nos primeiros testes C2500, o Mosquitto estabilizou exatamente em aproximadamente 1017 conexões nas três repetições.

A investigação mostrou:

```text
systemctl:
LimitNOFILE=524288

processo Mosquitto:
Max open files = 1024 soft / 524288 hard
```

No EMQX:

```text
Max open files = 1048576 / 1048576
```

O valor de aproximadamente 1017 conexões era compatível com um limite soft de 1024 descritores, considerando os descritores já utilizados pelo próprio processo.

Portanto, o primeiro resultado C2500 do Mosquitto **não representava a capacidade do broker**, mas uma limitação externa do processo.

O limite foi equalizado para:

```text
1048576 / 1048576
```

Após a correção, o Mosquitto passou a estabelecer aproximadamente 2500 conexões normalmente.

### 12.1 Decisão experimental

Os seis testes C2500 foram repetidos:

- 3 EMQX;
- 3 Mosquitto.

O coletor também foi ampliado de 90 s para **120 s**, permitindo registrar integralmente os 60 s de hold.

Os C2500 corrigidos substituem os originais nos resultados oficiais.

Os C2500 anteriores à correção permanecem como registros diagnósticos e estão excluídos das médias oficiais.

Não foi necessário repetir C500 e C1000, pois esses cenários não atingiram o limite de 1024 descritores.

---

## 13. Principais conclusões desta fase

Os resultados desta etapa permitem registrar as seguintes conclusões.

### 13.1 Capacidade básica de entrega

Em cargas abaixo do limite observado de escoamento, ambos os brokers apresentaram comportamento adequado.

Exemplos:

- C100/Q1: 100% de entrega observada;
- C500/Q0: aproximadamente 100% de entrega observada.

### 13.2 Efeito do QoS

O aumento do nível de QoS foi acompanhado de redução do throughput de recebimento.

No C500:

- QoS 0 manteve aproximadamente 500 msg/s;
- QoS 1 reduziu significativamente o throughput;
- QoS 2 apresentou nova redução.

No ambiente e nas configurações adotadas, o EMQX apresentou maior throughput recebido nos cenários QoS 1 e QoS 2.

### 13.3 Recursos

O Mosquitto apresentou menor consumo de:

- memória;
- CPU.

Essa característica é relevante para cenários com hardware restrito ou ambientes em que o baixo consumo seja o principal critério.

### 13.4 Volume de conexões

Depois da normalização do limite de descritores, ambos os brokers conseguiram aproximadamente 2500 conexões simultâneas com taxa de sucesso próxima de 100%.

Assim, os dados não sustentam a afirmação de que o EMQX “suportou 2500 conexões e o Mosquitto somente 1017”.

A limitação de 1017 observada inicialmente era consequência da configuração do sistema operacional/processo.

### 13.5 Escolha do EMQX

Os resultados de benchmark não indicam que o EMQX é superior ao Mosquitto em todos os critérios.

O Mosquitto apresentou vantagens importantes em consumo de recursos e bom comportamento nos testes de conexão.

A escolha tecnológica da plataforma considera os resultados experimentais e o conjunto de requisitos do projeto, incluindo:

- autenticação;
- autorização;
- gerenciamento de clientes;
- organização e controle de tópicos;
- observabilidade;
- administração;
- integração com a futura plataforma IoT;
- adequação ao cenário da empresa.

CPU, RAM e latência compõem a avaliação junto aos requisitos arquiteturais e funcionais.

---

## 14. Limitações do experimento

As condições de rede, a arquitetura do consumidor e os recursos das instâncias delimitam o alcance dos resultados.

### 14.1 Rede pública

O gerador de carga executou fora da AWS e a comunicação ocorreu pela Internet.

Portanto:

- RTT;
- jitter;
- condições do provedor;
- rota entre cliente e AWS;

podem afetar a latência observada.

### 14.2 Latência end-to-end

A latência registrada não representa exclusivamente o tempo interno do broker.

Ela representa o caminho:

```text
publicador no Windows
        ↓
Internet
        ↓
broker na EC2
        ↓
Internet
        ↓
coletor MQTT no Windows
```

### 14.3 Consumidor único

O benchmark utiliza um consumidor MQTT independente único.

Nos cenários de maior carga com QoS 1/2, essa arquitetura de coleta participa do limite observado de throughput.

### 14.4 Configuração dos brokers

Os resultados representam os brokers **nas configurações adotadas no ambiente experimental**.

Valores padrão e mecanismos internos de fila, inflight e controle de fluxo podem ser diferentes entre as implementações.

O alcance dos resultados se restringe às configurações e à arquitetura avaliadas.

### 14.5 Infraestrutura

Os resultados representam instâncias EC2 t3.small. Outras máquinas e arquiteturas podem apresentar capacidades diferentes.
