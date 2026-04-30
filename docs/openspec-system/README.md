# OpenSpec System Documentation

## Scope
本目录用于沉淀当前仓库已实现系统的逆向文档。内容以实际代码为准，不补写未在仓库中观察到的需求，也不将外部依赖误写为本仓库内部实现。

## Canonical Entry Documents
本轮整编后，以下根目录文档是当前最优先阅读的统一入口：

- `system_overview.md`
- `domain_map.md`
- `routing_overview.md`
- `architecture_notes.md`
- `shared_infrastructure.md`
- `frontend_backend_mapping.md`
- `observability_data_sources.md`
- `source_governance.md`
- `module_index.md`
- `file_to_module_map.md`
- `route_to_module_map.md`
- `api_to_module_map.md`
- `page_inventory.md`
- `api_inventory.md`
- `model_inventory.md`

这些文件用于统一命名、统一边界和统一索引。

## Supporting Documents
以下子目录文档仍然保留，但其角色已经调整为“按主题展开的支撑文档”，不再作为唯一入口：

- `system/`：较早阶段的系统说明与专项拆解。
- `modules/`：13 个业务模块的详细 OpenSpec 文档。
- `shared-infrastructure/`：共享基础设施专题。
- `cross-cutting/`：权限、license、数据范围、外部边界等横切主题。
- `source-mapping/`：源码与模块映射的工作底稿。
- `analysis/`：分析顺序、开放问题、漂移记录。

## Structure
- 根目录索引文档：统一入口、清单和映射。
- `modules/`：业务模块细节，不重复展开全局基础设施。
- `shared-infrastructure/`：前后端共享设施。
- `cross-cutting/`：跨业务域的横切关注点。
- `analysis/`：未决问题与核查记录。

## Notes
后续补全文档时，应优先把前端 route、前端 API client、后端 route group、service、repository、外部系统依赖串成统一链路，而不是只按目录逐个解释文件。模块文档中涉及共享壳层、权限模型、license、数据范围、外部边界时，应引用根目录索引或专题文档，而不是在每个模块内重复解释。
