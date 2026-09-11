# My Pet

一个使用 [Ebitengine](https://ebitengine.org/) 绘制的桌面小宠物示例。

## 下载

到 [GitHub Releases](https://github.com/fengxuan/MyPet/releases/latest) 下载对应系统的 zip，解压后保持 `assets` 和可执行文件在同一目录：

| 系统 | 文件 |
| --- | --- |
| macOS Apple 芯片 | `mypet-*-darwin-arm64.zip` |
| macOS Intel | `mypet-*-darwin-amd64.zip` |
| Windows 64 位 | `mypet-*-windows-amd64.zip` |
| Windows ARM | `mypet-*-windows-arm64.zip` |
| Linux x64 | `mypet-*-linux-amd64.zip` |
| Linux ARM64 | `mypet-*-linux-arm64.zip` |

发布新版本：打上 `vX.Y.Z` 标签并 push，GitHub Actions 会自动编译并挂到 Release。

```bash
git tag v0.1.0
git push origin v0.1.0
```

## 运行

```bash
go mod tidy
go run .
```

## 构建

```bash
./build.sh          # 打包全部可交叉编译的桌面平台
./build.sh local    # 只构建当前电脑
./build.sh windows  # 只打 Windows
./dist/mypet        # 当前平台可直接运行
```

每个成功的平台会生成：

- `dist/mypet-<os>-<arch>/`：可执行文件 + `assets`
- `dist/mypet-<os>-<arch>.zip`：可分发的压缩包

Windows / macOS 可在本机交叉编译。Linux 需要 C 编译器（`zig` 或 `x86_64-linux-gnu-gcc`），没有则自动跳过。解压后请保持 `assets` 和可执行文件在同一目录。正式发布走 GitHub Release，不必把 `dist/` 提交进仓库。

程序现在使用透明、无边框、置顶窗口，只显示宠物本身，不会显示后面的背景面板。

把鼠标移动到小猫身上，它会睁眼、举起爪子并挥动尾巴；移开鼠标后会恢复成安静的待机动作。点击小猫或按空格会播放拍打电脑动作，按住鼠标左键可以拖动宠物。在小猫上点右键会弹出「下一只」，再点一下会打开宠物选择框：默认显示 5 只，滚轮可以翻看，点缩略图切换。按 `Esc` 可以退出。

动画资源放在 [assets/pets](/Users/jianfengxuan/Documents/project/MyPet/assets/pets) 里。每个子文件夹都是一只动物，把自定义 PNG 丢进去后，右键「下一只」就能切到。也支持直接放入 Codex 的 `pet.json` + `spritesheet.webp` 宠物包，或自动读取 `~/.codex/pets/`。当前自带 Imagine 橘猫、奶油猫 tuft，以及 Pet Cats Pack 里的 6 只像素猫。

动作快慢在 [assets/config.txt](/Users/jianfengxuan/Documents/project/MyPet/assets/config.txt) 里改。格式不对或某一项写错时，系统只用默认值，不会崩。

## 交互逻辑

- 使用椭圆范围判断鼠标是否悬停在宠物身上。
- 待机状态：小猫闭眼、爪子下垂、尾巴缓慢摆动。
- 悬停状态：小猫睁眼、爪子抬起、头顶出现注意力提示。
- 拖动状态：冻结当前画面，只移动透明窗口，避免拖动时画面闪烁。
- 右键弹出「下一只」，点击后打开可滚动的宠物选择框（默认 5 只）。`assets/pets/` 里新放的动物文件夹、以及 `~/.codex/pets/` 里的 Codex 宠物会在右键时重新扫描。
