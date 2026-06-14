# 上下文工程 (Context Engineering) 前沿调研报告

> **调研范围**: 2025–2026 年，来自 MIT、Stanford、Microsoft、PKU、Tsinghua、ICLR/NeurIPS/ICML 等顶会与权威机构
> **调研日期**: 2026-06-15
> **总论文数**: 15+

---

## 一、上下文工程的定义与学科确立

### 1.1 **Context Engineering 综述**（2025年7月）

- **来源**: [A Survey of Context Engineering for Large Language Models](https://ar5iv.labs.arxiv.org/html/2507.13334v2) (arXiv:2507.13334)
- **机构**: 中科院、UC Merced、昆士兰大学、北大、清华
- **核心贡献**: **首篇正式定义上下文工程为独立学科的综述**，分析 ~1400 篇论文

**方法论与创新点**:

1. **形式化定义**: 上下文工程 = 对注入 LLM 的信息载荷进行系统优化的多目标优化问题
   ```
   C = A(c1, c2, ..., cn)  →  最优上下文 = 组装函数(指令, 知识, 工具, 记忆, 状态, 查询)
   ```

2. **双层分类法**:
   - **基础组件层**: 上下文检索与生成 → 上下文处理 → 上下文管理
   - **系统实现层**: RAG(模块化/Agentic/图增强) → 记忆系统 → 工具集成推理 → 多Agent系统

3. **关键发现**: 存在 **理解-生成不对称性**——模型擅长理解复杂上下文，但生成同样复杂的输出能力很差。这被确定为优先研究方向

4. **解决的问题**: 将上下文工程从零散的 prompt engineering 提升为系统化学科，给出统一技术路线图

---

## 二、MIT 核心成果

### 2.1 **MEM1: 用强化学习协同记忆与推理**（ICLR 2026）

- **来源**: [MEM1: Learning to Synergize Memory and Reasoning for Efficient Long-Horizon Agents](https://iclr.cc/virtual/2026/poster/10008961)
- **机构**: MIT (Daniela Rus, Paul Liang 等)
- **代码**: https://github.com/MIT-MI/MEM1

**方法论与创新点**:

1. **端到端 RL 框架**: 用强化学习训练代理维护一个紧凑的内部状态，该状态同时支持记忆巩固和推理
2. **近乎恒定的上下文大小**: 在解决长程任务时不无限增长上下文，而是通过 RL 学习丢弃无关信息
3. **Rollout轨迹截断**: 训练中截断长轨迹，使内化状态学会整合新旧观察

**效果**: 相比 Qwen2.5-14B-Instruct，MEM1-7B 性能提升 **3.5x**，内存减少 **3.7x**，且泛化到训练时未见的时间跨度

**解决的问题**: 长程对话中上下文无限膨胀问题，同时保持决策质量

---

### 2.2 **递归语言模型 RLM: 用代码管理上下文**（MIT 2025）

- **来源**: [Recursive Language Models](https://introl.com/blog/recursive-language-models-rlm-context-management-2026) (arXiv:2512.24601)
- **机构**: MIT

**方法论与创新点**:

1. **"上下文折叠"**: 不把所有内容塞进上下文，而是教模型使用 **Python 脚本 + 子 LLM 调用** 主动管理上下文
2. **持久化 Python REPL**: 主模型将任务委托给一个持久化的 Python 运行时和可生成的子 LLM 实例
3. **分层委托架构**: 主模型只处理高层次决策，子 LLM 处理具体细节，结果折叠回上下文

**效果**:
- 处理超出模型上下文窗口 **100x** 的输入
- CodeQA: GPT-5 基线 24% → RLM **62%**
- Oolong(1.5M字符): 标准 LLM ~35% → RLM **~75%**
- 主模型 token 消耗减少 **2-3x**

**解决的问题**: 上下文窗口物理限制，将线性扩展问题转化为分层管理问题

---

### 2.3 **M+: 可扩展长期记忆**（ICML 2025）

- **来源**: UCSD / MIT-IBM Watson AI Lab
- **链接**: ICML 2025 proceedings

**方法论与创新点**:

1. **压缩潜记忆**: 每层使用 256 个压缩向量（而不是 1024 个 KV pair）进行检索，降低检索开销
2. **协同训练的检索器**: 检索器与记忆机制联合训练，直接优化端到端任务性能
3. **与 KV 缓存集成**: 压缩的潜记忆直接注入注意力层

**效果**: 知识保留从 **<20K → >160K tokens**，GPU 内存开销相似

**解决的问题**: 长期知识保留与推理效率的权衡

---

## 三、上下文压缩前沿技术

### 3.1 **COMI: 基于边际信息增益的粗到细压缩**（ICLR 2026）

- **来源**: [COMI: Coarse-to-fine Context Compression via Marginal Information Gain](https://iclr.cc/virtual/2026/poster/10009801)
- **机构**: ICLR 2026 接收

**方法论与创新点**:

1. **边际信息增益 (MIG)**: 一个全新的压缩指导指标，衡量一个单元的**相关性减去其与其他单元的语义冗余**，确保压缩保留既相关又不冗余的信息
2. **两阶段压缩**:
   - 粗粒度: 动态分组分配不同压缩率
   - 细粒度: 组内基于 MIG 权重的 token 合并
3. **非暴力压缩**: 不是简单丢弃 token，而是**合并语义相似的 token**

**效果**: **32x 压缩率**下，Qwen2-7B 在 NaturalQuestions 上提升约 25 points Exact Match

**解决的问题**: 高压缩率下信息丢失的关键问题——通过 MIG 保留"独特信息"而非冗余信息

---

### 3.2 **Parallel Context Compaction: 并行上下文压缩**（2026年5月）

- **来源**: [Parallel Context Compaction for Long-Horizon LLM Agent Serving](https://export.arxiv.org/abs/2605.23296)

**方法论与创新点**:

1. **并行压缩**: 压缩与 Agent 推理**并发运行**，不阻塞主流程
2. **可预测的压缩量控制**: 操作者可以精细控制摘要体积
3. **跨架构验证**: 在 8B–120B 参数模型上验证（dense、MoE、推理、非推理架构）

**效果**: 端到端耗时大幅降低，在 HotpotQA 和 LoCoMo 上验证

**解决的问题**: 传统顺序摘要式压缩导致推理停滞数十秒的问题

---

### 3.3 **RAM: 类人"精读+略读"压缩**（2026年2月）

- **来源**: [Read As Human: Compressing Context via Parallelizable Close Reading and Skimming](https://browse-export.arxiv.org/abs/2602.01840)

**方法论与创新点**:

1. **模拟人类阅读模式**: 高相关段落 **精读**（全文保留），低相关段落 **略读**（压缩为紧凑摘要向量）
2. **并行编码**: 上下文段与输入查询同时编码
3. **对比学习**: 用对比学习目标优化精读/略读的分界决策

**效果**: 最长 32K token 输入上实现最高 **12x 端到端加速**，同时保持 QA 和摘要性能

**解决的问题**: 长上下文的计算成本呈线性增长，通过选择性压缩降低到亚线性

---

### 3.4 **ReSum: 强化学习协同推理与摘要**（2026年6月）

- **来源**: [ReSum: Synergizing LLM Reasoning and Summarization with Reinforcement Learning](https://export.arxiv.org/abs/2606.13316)

**方法论与创新点**:

1. **自我摘要**: LLM 在推理过程中自发地压缩和组织自己的推理轨迹
2. **RLVR + 摘要**: 在带有可验证奖励的强化学习框架中嵌入摘要机制，防止无限长推理
3. **对比评估**: 用对比方法评估摘要质量，学习何时需要自我摘要

**效果**: 任务性能提升约 **4%**，同时**推理长度缩减 18.6%**

**解决的问题**: RL 训练中模型倾向于产生越来越长的推理链，消耗上下文预算

---

### 3.5 **AdmTree: 自适应语义树压缩**（NeurIPS 2025）

- **来源**: [AdmTree: Compressing Lengthy Context with Adaptive Semantic Trees](https://proceedings.neurips.cc/paper_files/paper/2025/hash/39a397e1721555ddfa83cea2de225760-Abstract-Conference.html)

**方法论与创新点**:

1. **语义二叉树**: 基于信息密度动态分段，用**摘要 token** 摘要变长段作为二叉树的叶子节点
2. **轻量聚合**: 轻量级聚合机制 + 冻结骨干 LLM，实现高效层次化摘要
3. **无位置偏见**: 通过树结构避免了 lost-in-the-middle 问题

**效果**: 同时保留细节信息和全局语义连贯性，减轻位置偏见

**解决的问题**: LLM 的 U 型注意力曲线——中间位置信息丢失（lost-in-the-middle）

---

## 四、Agent 驱动的主动上下文管理

### 4.1 **Active Context Compression: Agent 自主记忆管理**（2026年1月）

- **来源**: [Active Context Compression via Agent-Controlled Focus Primitives](https://arxiv.org/abs/2601.07190)
- **机构**: Zeph 项目实现

**方法论与创新点**:

1. **Agent 自主决策**: Agent 自主决定何时将探索历史合并为持久的 **Knowledge Block**
2. **聚焦工具**: 两个工具 `start_focus(scope)` 和 `complete_focus(summary)` 让 Agent 自己管理压缩
3. **锯齿形上下文模式**: 上下文在探索期增长 → 在合并期坍缩，形成锯齿形曲线

**效果**: SWE-bench Lite 上 token 减少 **22.7%**（14.9M → 11.5M），**准确率零损失**，每次任务平均压缩 6 次

**解决的问题**: 将"何时压缩"的决策权交给 Agent 本身，而不是依赖人为预设的阈值

---

### 4.2 **Agentic Context Engineering (ACE)**（ICLR 2026）

- **来源**: Stanford / SambaNova / UC Berkeley
- **工具**: 已有 `agent_episodic_memory` LangChain 中间件

**方法论与创新点**:

1. **Generator/Reflector/Curator 架构**: 将 Agent 上下文视为不断进化的"剧本"
2. **Delta 更新**: 跨运行仅做增量更新（不是整体重写），大幅降低开销
3. **回放与精炼**: 通过回放历史运行，精炼上下文表示

**效果**: AppWorld 上 **+10.6%**，适应延迟降低 **86.9%**

**解决的问题**: Agent 上下文的冷启动问题和跨运行知识复用

---

### 4.3 **MEM1** 与 **Memex(RL)** 对比

| 方面 | MEM1 (MIT, ICLR 2026) | Memex(RL) (2026.04) |
|------|----------------------|---------------------|
| 方法 | 紧凑 RL 内化状态 | 索引式经验记忆 + 外部 KV 存储 |
| 压缩方式 | 端到端学习丢弃无关信息 | Agent 学习何时压缩/归档/索引/解引用 |
| 上下文 | 近乎恒定大小 | 紧凑在线摘要 + 全保真外部存储 |
| 理论 | 实验验证 | 提供理论分析，证明有界工作上下文可保持决策质量 |

---

## 五、蒸馏式上下文方法

### 5.1 **Context Distillation as Latent Memory Management**（2026年5月）

- **来源**: [Context Distillation as Latent Memory Management](https://export.arxiv.org/abs/2605.28889)

**方法论与创新点**:

1. **LoRA 适配器记忆库**: 每个上下文蒸馏为独立的 LoRA 适配器，形成模块化记忆银行
2. **Self-Gating 机制**: 自门控决定是否激活潜记忆
3. **缓存共享**: 多个适配器共享计算缓存，降低推理开销

**解决的问题**: 单一模型难以存储所有上下文，模块化适配器实现"即插即用"的知识注入

---

### 5.2 **Doc-to-LoRA: 单次前向传播内化上下文**（ICML 2026）

- **来源**: ICML 2026, Charakorn, Cetin, Uesaka, Lange

**方法论与创新点**:

1. **轻量级超网络**: 元学习对任意未见提示进行近似上下文蒸馏
2. **单步生成**: 给定未见过的提示，D2L 即时生成 LoRA 适配器
3. **无需重新消费**: 查询时无需重新处理原始上下文

**效果**: 在 needle-in-a-haystack 任务上，在超出模型原生上下文窗口 **4x** 时仍接近完美准确率

**解决的问题**: 传统上下文压缩需要重新处理全部历史，D2L 做到"一次消化，多次查询"

---

## 六、企业级与实用化方案

### 6.1 **Less Context, Better Agents**（Microsoft, 2026年6月）

- **来源**: [Less Context, Better Agents: Efficient Context Engineering for Long-Horizon Tool-Using LLM Agents](https://browse-export.arxiv.org/abs/2606.10209)

**方法论与创新点**:

1. **选择性保留**: 仅保留最近 N 次工具调用，丢弃早期交互
2. **自动摘要**: 对丢弃的部分做自动化摘要
3. **实证验证**: 在 Microsoft Dynamics 365 Finance 的企业工作流上验证

**效果**: 保留最近 5 次工具调用 + 摘要 → **91.6% 完成率**（完整上下文 71%，无上下文 8%），token 减少 **63%**，耗时减少 **60%**

**解决的问题**: 企业场景下上下文不是越长越好，选择性压缩反而提升效果

---

### 6.2 **Context Cartography: 上下文空间治理框架**（2026年3月）

- **来源**: [Context Cartography: Toward Structured Governance of Contextual Space](https://ar5iv.labs.arxiv.org/html/2603.20578)

**方法论与创新点**:

1. **三区模型**: 黑雾（未观测区）、灰雾（存储记忆）、可视区（主动推理）
2. **7 个制图算子**: 侦察、选择、简化、聚合、投影、置换、分层
3. **跨系统分析**: 用此框架分析 Claude Code、Letta、MemOS、OpenViking

**解决的问题**: 缺乏统一理论框架来比较和理解不同的上下文管理系统

---

## 七、总结：核心趋势与对 agent-core 的启示

### 7.1 六大趋势

| 趋势 | 代表工作 | 对库的启示 |
|------|---------|-----------|
| **并行化** | Parallel Compaction, CoMem | 压缩不应阻塞推理，应异步/并行 |
| **Agent 自主决策** | ACE, Active Context Compress, MEM1 | 阈值不应硬编码，Agent 应自主决策 |
| **混合层级压缩** | COMI, AdmTree, RAM | 粗粒度(段级) + 细粒度(token级)联合 |
| **RL 优化记忆** | MEM1, Memex(RL), ReSum | RL 可训练出更好的记忆管理策略 |
| **蒸馏式存储** | Doc-to-LoRA, Context Distillation | 压缩为 LoRA/适配器而不是纯文本 |
| **企业简洁方案** | Less Context Better Agents | 有时少即是多，选择性保留胜过全保留 |

### 7.2 对当前 `agent-core` 库的改进方向

1. **压缩策略**: 当前仅支持 summary/mermaid/sliding，可考虑
   - 异步并行压缩（Parallel Compaction 模式）
   - 多层级压缩（粗粒度 + 细粒度）
   - MIG（边际信息增益）作为压缩质量指标

2. **卸载决策**: 当前固定 50%/85% 阈值，可考虑
   - RL 训练自适应阈值
   - ACE 式 Generator/Reflector/Curator 架构
   - Agent 自主控制（Active Context Compression）

3. **记忆系统**: 当前 L0-L3 线性流水线，可考虑
   - LoRA 适配器记忆库（Context Distillation）
   - 索引式经验记忆（Memex(RL)）
   - 语义树存储（AdmTree）

---

## 参考文献

1. [A Survey of Context Engineering for Large Language Models](https://ar5iv.labs.arxiv.org/html/2507.13334v2) — 中科院/北大/清华, 2025
2. [MEM1: Learning to Synergize Memory and Reasoning](https://iclr.cc/virtual/2026/poster/10008961) — MIT, ICLR 2026
3. [Recursive Language Models](https://introl.com/blog/recursive-language-models-rlm-context-management-2026) — MIT, 2025
4. [M+: Extending MemoryLLM with Scalable Long-Term Memory](https://cseweb.ucsd.edu/~jmcauley/reviews/icml25a.pdf) — UCSD/MIT-IBM, ICML 2025
5. [COMI: Coarse-to-fine Context Compression via Marginal Information Gain](https://iclr.cc/virtual/2026/poster/10009801) — ICLR 2026
6. [Parallel Context Compaction for Long-Horizon LLM Agent Serving](https://export.arxiv.org/abs/2605.23296) — 2026
7. [Read As Human: Compressing Context via Parallelizable Close Reading and Skimming](https://browse-export.arxiv.org/abs/2602.01840) — 2026
8. [ReSum: Synergizing LLM Reasoning and Summarization with Reinforcement Learning](https://export.arxiv.org/abs/2606.13316) — 2026
9. [AdmTree: Compressing Lengthy Context with Adaptive Semantic Trees](https://proceedings.neurips.cc/paper_files/paper/2025/hash/39a397e1721555ddfa83cea2de225760-Abstract-Conference.html) — NeurIPS 2025
10. [Active Context Compression: Autonomous Memory Management in LLM Agents](https://arxiv.org/abs/2601.07190) — 2026
11. [Agentic Context Engineering (ACE)](https://conf.researchr.org/details/icse-2026/llm4code-2026-papers/33/ContextPilot-Code-Context-Engineering-with-Memory-Augmented-Exploration-Agents) — Stanford/SambaNova/Berkeley, ICLR 2026
12. [Context Distillation as Latent Memory Management](https://export.arxiv.org/abs/2605.28889) — 2026
13. [Doc-to-LoRA: Learning to Instantly Internalize Contexts](https://icml.cc/virtual/2026/poster/62227) — ICML 2026
14. [Less Context, Better Agents: Efficient Context Engineering for Long-Horizon Tool-Using LLM Agents](https://browse-export.arxiv.org/abs/2606.10209) — Microsoft, 2026
15. [Context Cartography: Toward Structured Governance of Contextual Space](https://ar5iv.labs.arxiv.org/html/2603.20578) — 2026
16. [Memex(RL): Scaling Long-Horizon LLM Agents via Indexed Experience Memory](https://ar5iv.labs.arxiv.org/html/2603.04257) — 2026
