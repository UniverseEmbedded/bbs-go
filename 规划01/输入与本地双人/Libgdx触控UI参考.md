# Libgdx触控UI参考

本文档整理Libgdx版Duel的Android触控UI/UX实现，作为移动端和VR输入设计的参考。

---

## 核心设计原则

### 1) 移动输入用"方向优先、幅度次要"

归一化dx/dy，速度交给物理/惯性系统。

**源码依据：**

```java
if(mag<0.01f) { dx=0; dy=0; }
else { dx=dxIn/mag; dy=dyIn/mag; }
```

玩家手指不需要精确推到某个距离，只要方向正确就能玩，学习成本低。

### 2) 按钮/手柄只做"输入翻译层"

触控/VR手柄 → 统一InputData（或keyDown/up）

玩法层永远只读一种输入结构，未来加联机/回放/重映射会省很多命。

### 3) UI极简 + 调试隔离 + 适配条件化

- 正式版只保留最关键2–3个控件
- debug功能只在debug开关下出现
- 布局与启用由条件控制

---

## 摇杆实现

### 触点锚定（Anchored-at-Touch）

**不是固定摇杆，也不是双圆盘。**

- 玩家在**左半屏（横屏）**或**除底部按钮区以外的屏幕区域（竖屏）**任意点按下
- **按下那一刻的触点**就成为"摇杆中心/原点"
- 手指拖动时，系统计算"当前触点 - 原点"的向量作为dx/dy
- 松手时清零

### 区域分割（activeCondition）

**横屏默认：**`osx < width/2` → 左半屏才会创建移动控制

**竖屏默认：**`osy < height - bu*2` → 底部按钮区不允许触发摇杆

```java
public static final GetBooleanWithInfo landscapeCondition=(p,info)->info.osx<p.width/2f,
  portraitCondition=(p,info)->info.osy<p.height-p.bu*2;
```

**只认第一根手指：**

```java
if(active&&activeCondition.get(p,info)) {
  if(moveCtrl==null) moveCtrl=info;
}
```

### 摇杆UI（极简叠加层）

**画三件事：**

1. 在原点画一个十字（cross）
2. 从原点到当前触点画一条线
3. 在当前触点画一个小十字
4. 另外还画了一个arc（弧线），方向来自atan2(dx,dy)

```java
p.cross(moveCtrl.sx,moveCtrl.sy,32*scale,32*scale);   // 原点十字
p.line(moveCtrl.x,moveCtrl.y,moveCtrl.sx,moveCtrl.sy);// 拖拽线
p.cross(moveCtrl.x,moveCtrl.y,16*scale,16*scale);     // 当前点十字
float deg=UtilMath.deg(UtilMath.atan2(dxCache,dyCache));
p.arc(moveCtrl.sx,moveCtrl.sy,magCache,45-deg,90);    // 方向/强度提示弧
```

**结论：**libgdx版的"摇杆UI"本质是"触点可视化"，完全不依赖一个固定摆在左下角的摇杆控件。

### 输入向量计算

```java
dxCache=moveCtrl.x-moveCtrl.sx;
dyCache=moveCtrl.y-moveCtrl.sy;
magCache=min(mag(dxCache,dyCache), maxDist);
targetTouchMoved(dxCache,dyCache, magCache);
```

**maxDist限制：**

```java
public void updateMaxDist() { maxDist=p.u*scale; }
public void frameResized(int w,int h) { updateMaxDist(); }
```

### 世界坐标 vs 屏幕坐标

**AndroidCtrlBase：**用世界坐标（x/sx），会受相机变化影响

**AndroidCtrlScreenBase：**用屏幕坐标（ox/osx），不受相机影响

```java
dxCache=moveCtrl.ox-moveCtrl.osx;
dyCache=moveCtrl.oy-moveCtrl.osy;
```

**重要：**如果相机/画布会缩放/平移，屏幕坐标版更不容易出现"我手指没动但输入漂了"的怪体验。

---

## 按钮实现

### 攻击按钮（Z / X）

**直接模拟键盘事件：**

```java
p.inputProcessor.keyDown(Input.Keys.Z);
p.inputProcessor.keyUp(Input.Keys.Z);

p.inputProcessor.keyDown(Input.Keys.X);
p.inputProcessor.keyUp(Input.Keys.X);
```

**布局：**右下角两枚大按钮

- Z在`p.width - p.bu*4f`
- X在`p.width - p.bu*2.5f`
- y都在`p.height - p.bu*1.5f`

**UX优点：**

- 按钮大、位置固定、肌肉记忆强
- 最关键：它不改游戏内输入体系，只是把触控"翻译成键盘事件"

### Debug按钮隔离

```java
if(pg.p.debug) return concat(superButtons,
  new TextButton<>(" C") ... keyDown(Input.Keys.C) ... keyUp(Input.Keys.C)
);
else return superButtons;
```

---

## 横竖屏适配

```java
if(p.isAndroid) {
  actrl=new DuelAndroidCtrl(p,this);
  if(p.config.data.orientation==1) actrl.activeCondition=AndroidCtrlBase.portraitCondition;
  else actrl.activeCondition=AndroidCtrlBase.landscapeCondition;
  actrl.init();
}
```

控制UI不是"写死布局"，而是**跟随横竖屏条件启用/布局策略**。

---

## 对VR设计的启发

### 输入翻译层

如果做"左手移动/右手普攻大招"，也可以学这招：

- 不要把VR输入逻辑散落在玩法代码里
- 做一层"输入翻译器"（VR手柄 → 统一的InputData）
- 未来加联机、AI、回放、重映射都会轻松很多

### 屏幕坐标优先

如果做"水箱观察窗口+相机可能会轻微跟随/缩放/抖动"：

- 控制输入最好绑定到屏幕/手柄局部空间，而不是世界空间
- 否则会出现"我没动但输入变了"的灾难体验

### 条件驱动的控制层

如果后面考虑Quest/Pico等设备的不同手柄布局、左右手互换、坐姿/躺姿模式：

- 建议做成"条件驱动的控制层配置"
- 不要硬编码一套
