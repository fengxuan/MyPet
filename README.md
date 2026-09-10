# My Pet

一个使用 [Ebitengine](https://ebitengine.org/) 绘制的桌面小宠物示例。

## 运行

```bash
go mod tidy
go run .
```

## 构建

```bash
./build.sh
./dist/mypet
```

构建脚本会在 `dist/mypet` 生成当前平台的可执行文件。

程序现在使用透明、无边框、置顶窗口，只显示宠物本身，不会显示后面的背景面板。

把鼠标移动到小猫身上，它会睁眼、举起爪子并挥动尾巴；移开鼠标后会恢复成安静的待机动作。点击小猫或按空格会播放拍打电脑动作，按住鼠标左键可以拖动宠物。在小猫上点右键会弹出「下一只」，再点一下就能换成下一只猫。按 `Esc` 可以退出。

动画资源放在 [assets/pets](/Users/jianfengxuan/Documents/project/MyPet/assets/pets) 里。每个子文件夹都是一只动物，把自定义 PNG 丢进去后，右键「下一只」就能切到。当前自带 Imagine 橘猫和 Pet Cats Pack 里的 6 只像素猫。

动作快慢在 [assets/config.txt](/Users/jianfengxuan/Documents/project/MyPet/assets/config.txt) 里改。格式不对或某一项写错时，系统只用默认值，不会崩。

## 交互逻辑

- 使用椭圆范围判断鼠标是否悬停在宠物身上。
- 待机状态：小猫闭眼、爪子下垂、尾巴缓慢摆动。
- 悬停状态：小猫睁眼、爪子抬起、头顶出现注意力提示。
- 拖动状态：冻结当前画面，只移动透明窗口，避免拖动时画面闪烁。
- 右键弹出「下一只」，点击后切换到下一只。`assets/pets/` 里新放的动物文件夹会在右键时重新扫描，可循环切换。
