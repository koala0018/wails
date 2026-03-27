import { useMemo, useState } from "react";
import "./App.css";
import { Greet } from "../wailsjs/go/main/App";

const featureCards = [
    {
        title: "React + Wails",
        text: "前端页面和桌面壳已经联通，现在就可以继续搭业务界面。",
    },
    {
        title: "Go Backend Ready",
        text: "Go、Wails CLI 和桌面运行时都已经安装完成，后端方法可直接暴露给前端。",
    },
    {
        title: "Build Verified",
        text: "当前项目已经完成过真实构建验证，打开后就能直接看到效果。",
    },
];

const quickSteps = [
    "运行 wails dev 进入桌面开发模式",
    "在 frontend/src 下继续拆分页面和组件",
    "在 Go 中添加业务方法并通过 Wails 暴露给前端",
];

function App() {
    const [name, setName] = useState("");
    const [resultText, setResultText] = useState("输入你的名字，然后点按钮测试 Go 后端调用。");
    const [isLoading, setIsLoading] = useState(false);

    const greetingName = useMemo(() => name.trim() || "Developer", [name]);

    async function handleGreet() {
        setIsLoading(true);
        try {
            const result = await Greet(greetingName);
            setResultText(result);
        } finally {
            setIsLoading(false);
        }
    }

    return (
        <main className="app-shell">
            <section className="hero-card">
                <div className="hero-copy">
                    <span className="eyebrow">Desktop Starter</span>
                    <h1>你的桌面应用首页已经准备好了。</h1>
                    <p className="hero-text">
                        这不是默认模板页，而是一个适合作为开发起点的桌面应用首页。
                        启动应用后，你会立刻看到状态概览、开发建议，以及一个可直接验证
                        Go 后端联通的交互区。
                    </p>

                    <div className="hero-actions">
                        <button className="primary-btn" onClick={handleGreet} disabled={isLoading}>
                            {isLoading ? "连接中..." : "测试 Go 调用"}
                        </button>
                        <div className="status-pill">
                            <span className="status-dot" />
                            桌面应用已可直接预览
                        </div>
                    </div>
                </div>

                <div className="preview-panel">
                    <div className="window-bar">
                        <span />
                        <span />
                        <span />
                    </div>
                    <div className="preview-content">
                        <div className="preview-label">实时联调面板</div>
                        <div className="preview-result">{resultText}</div>
                        <label className="field-label" htmlFor="name">
                            你的名字
                        </label>
                        <input
                            id="name"
                            className="name-input"
                            value={name}
                            onChange={(event) => setName(event.target.value)}
                            autoComplete="off"
                            placeholder="例如：Zhuan"
                        />
                        <div className="hint-text">
                            这里会调用 Go 中的 <code>Greet()</code> 方法返回结果。
                        </div>
                    </div>
                </div>
            </section>

            <section className="content-grid">
                <div className="panel">
                    <div className="panel-title">开发起点</div>
                    <div className="feature-list">
                        {featureCards.map((item) => (
                            <article className="feature-card" key={item.title}>
                                <h2>{item.title}</h2>
                                <p>{item.text}</p>
                            </article>
                        ))}
                    </div>
                </div>

                <div className="panel side-panel">
                    <div className="panel-title">下一步建议</div>
                    <ol className="step-list">
                        {quickSteps.map((step) => (
                            <li key={step}>{step}</li>
                        ))}
                    </ol>

                    <div className="mini-metrics">
                        <div className="metric-card">
                            <strong>UI</strong>
                            <span>React + Vite</span>
                        </div>
                        <div className="metric-card">
                            <strong>Runtime</strong>
                            <span>Wails Desktop</span>
                        </div>
                        <div className="metric-card">
                            <strong>Backend</strong>
                            <span>Go Service Bridge</span>
                        </div>
                    </div>
                </div>
            </section>
        </main>
    );
}

export default App;
