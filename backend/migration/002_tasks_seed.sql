-- 任务演示数据：待办 4 条、进行中 3 条、已完成 2 条。
-- 截止日期相对于首次执行当天；由 Go 在首次初始化事务中执行。
-- 手动导入仅适用于已建表且无任务数据的库，重复导入会产生重复任务。
INSERT INTO tasks (title, description, assignee, priority, status, due_date, position) VALUES
  ('梳理商品发布流程', '明确新增、审核、上架之间的业务边界，并列出验收标准。', '林悦', 'high', 'todo', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL -1 DAY), '%Y-%m-%d'), 0),
  ('设计任务卡片组件', '将任务标题、优先级、负责人和截止日期整理为可复用的卡片。', '陈晨', 'medium', 'todo', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 2 DAY), '%Y-%m-%d'), 1),
  ('补充表单边界校验', '覆盖空标题、超长描述、非法日期和重复提交。', '周宁', 'high', 'todo', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 3 DAY), '%Y-%m-%d'), 2),
  ('整理交付清单', '核对课程源码、启动说明与接口契约。', '', 'low', 'todo', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 5 DAY), '%Y-%m-%d'), 3),
  ('实现跨列拖拽', '使用 useSortable 连接三列，处理 Vue 数据与 DOM 的同步。', '陈晨', 'high', 'doing', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 0 DAY), '%Y-%m-%d'), 4),
  ('联调任务编辑抽屉', '通过 Element Plus 完成回填、校验、保存和未保存提醒。', '林悦', 'medium', 'doing', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 1 DAY), '%Y-%m-%d'), 5),
  ('验证失败回滚', '模拟接口失败，确保卡片恢复原列和原顺序。', '周宁', 'medium', 'doing', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 2 DAY), '%Y-%m-%d'), 6),
  ('搭建 Gin 分层接口', '完成路由、参数校验、业务层和 GORM 仓储装配。', '周宁', 'high', 'done', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL -2 DAY), '%Y-%m-%d'), 7),
  ('接入后台导航', '复用 Vben 布局、菜单和现有登录能力。', '陈晨', 'low', 'done', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL -1 DAY), '%Y-%m-%d'), 8);
