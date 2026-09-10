# 动画资源目录

`pets/` 里每一个文件夹都是一只可切换的动物。放进去之后右键点「下一只」就能切到，不必改代码。缺的动作会自动用待机帧代替。

```text
assets/pets/兔子/          # 文件夹名字就是宠物名，随意
├── idle/                 # 推荐
│   ├── 000.png
│   └── 001.png
├── hover/                # 可省略
└── hit/                  # 可省略
```

也可以直接导入 [Codex Pet](https://github.com/openai/skills/blob/main/skills/.curated/hatch-pet/references/codex-pet-contract.md) 的文件夹（`pet.json` + `spritesheet.webp` / `.png`）。整夹拷进 `pets/` 即可，不必拆成 idle/hover/hit：

```text
assets/pets/codie/
├── pet.json
└── spritesheet.webp
```

已经装在 `~/.codex/pets/`（或 `$CODEX_HOME/pets/`）里的自定义宠物会自动出现在「下一只」里，不用再拷一份。同名文件夹以 `assets/pets/` 为准。

精灵图按 Codex 的 8 列 × 9 行（或 11 行）格子切开：待机用 idle，悬停用 waving，点击用 jumping（没有则用 failed）。

也可以更随便：

```text
assets/pets/兔子/
├── 000.png               # 只有这些时，全部当作待机
└── 001.png
```

或直接丢精灵图（横向/竖向一排帧，程序会按正方形切开）：

```text
assets/pets/兔子/
├── idle.png
├── meow.png              # 当作悬停
└── itch.png              # 当作点击
```

- `idle`（也认 stand / sleep / walk / sitting）：待机
- `hover`（也认 meow / alert / look）：鼠标悬停
- `hit`（也认 itch / lick / slap）：点击或空格

播放速度在 [config.txt](config.txt) 里改，默认待机 3 FPS、悬停 4 FPS、拍打 6 FPS，拍打持续 1.35 秒。某一项写错或超出范围时，只该项回到默认值。

嵌套目录也可以，例如把整包 `Pet Cats Pack/Cat-1` 拷进 `pets/`。以 `.` 开头的文件夹会忽略。

右键弹出「下一只」，点击后循环切换。当前动物按文件夹名字记住，下次启动还是它。

所有帧建议使用透明 PNG。程序按文件名排序并缩放到窗口内。小尺寸像素图会用最近邻放大，避免发糊。

如果目录没有 PNG，程序会自动使用内置绘制的小猫作为后备，不影响运行。
