# Android 桌面宠物

悬浮窗架构：

```text
PetService（前台 Service）
        ↓
  WindowManager
        ↓
   悬浮窗口
        ↓
    PetView
```

## 运行

1. 用 Android Studio 打开 `android/` 目录，或：

```bash
cd android
./gradlew :app:assembleDebug
```

2. 安装后打开 App，先授权 **悬浮窗**，再点「启动桌面宠物」。
3. 小猫会出现在其它 App 上面：
   - 拖动：移动
   - 点一下：拍打
   - 长按：打开宠物选择框（默认 5 只，可滚动）

宠物帧来自仓库根目录的 `assets/pets/`，和桌面版共用。
