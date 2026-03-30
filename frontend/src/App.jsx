import { useEffect, useState } from "react";
import "./App.css";
import {
    GetEnvironmentStatus,
    SelectOutputDirectory,
    SelectVideoFile,
    SplitVideo,
} from "../wailsjs/go/main/App";

const initialForm = {
    inputPath: "",
    outputDir: "",
    segmentLengthSec: 30,
};

function App() {
    const [form, setForm] = useState(initialForm);
    const [environment, setEnvironment] = useState({
        ffmpegReady: false,
        ffmpegPath: "",
        message: "正在检查 ffmpeg 环境...",
    });
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [errorMessage, setErrorMessage] = useState("");
    const [result, setResult] = useState(null);

    useEffect(() => {
        async function loadEnvironment() {
            try {
                const status = await GetEnvironmentStatus();
                setEnvironment(status);
            } catch (error) {
                setEnvironment({
                    ffmpegReady: false,
                    ffmpegPath: "",
                    message: `环境检查失败：${String(error)}`,
                });
            }
        }

        loadEnvironment();
    }, []);

    async function handleSelectVideo() {
        const selection = await SelectVideoFile();
        if (!selection?.path) {
            return;
        }

        setForm((current) => ({
            ...current,
            inputPath: selection.path,
        }));
        setErrorMessage("");
    }

    async function handleSelectOutputDir() {
        const selection = await SelectOutputDirectory();
        if (!selection?.path) {
            return;
        }

        setForm((current) => ({
            ...current,
            outputDir: selection.path,
        }));
        setErrorMessage("");
    }

    async function handleSubmit(event) {
        event.preventDefault();
        setIsSubmitting(true);
        setErrorMessage("");
        setResult(null);

        try {
            const response = await SplitVideo({
                inputPath: form.inputPath.trim(),
                outputDir: form.outputDir.trim(),
                segmentLengthSec: Number(form.segmentLengthSec),
            });
            setResult(response);
        } catch (error) {
            setErrorMessage(String(error));
        } finally {
            setIsSubmitting(false);
        }
    }

    return (
        <main className="page-shell">
            <section className="hero">
                <div className="hero-copy">
                    <div className="eyebrow">Desktop Video Splitter</div>
                    <h1>上传需要切割的视频</h1>
                    <p className="hero-text">
                        这是一个本地桌面工具，直接调用 ffmpeg 处理视频，不走云端。
                        你只需要选择源视频、设置每段秒数、指定保存目录，然后点击开始切割。
                    </p>

                    <div className="status-grid">
                        <article className={`status-card ${environment.ffmpegReady ? "ready" : "warning"}`}>
                            <span className="status-label">ffmpeg 环境</span>
                            <strong>{environment.ffmpegReady ? "已就绪" : "未就绪"}</strong>
                            <p>{environment.message}</p>
                        </article>

                        <article className="status-card">
                            <span className="status-label">切割方式</span>
                            <strong>自动分段导出</strong>
                            <p>输出为 MP4，按设定秒数连续生成多个短视频文件。</p>
                        </article>
                    </div>
                </div>

                <form className="tool-card" onSubmit={handleSubmit}>
                    <div className="section-heading">
                        <span>操作面板</span>
                        <strong>视频切割设置</strong>
                    </div>

                    <label className="field-block">
                        <span>源视频</span>
                        <div className="path-row">
                            <input
                                value={form.inputPath}
                                onChange={(event) =>
                                    setForm((current) => ({
                                        ...current,
                                        inputPath: event.target.value,
                                    }))
                                }
                                placeholder="选择或粘贴视频文件路径"
                            />
                            <button type="button" className="secondary-btn" onClick={handleSelectVideo}>
                                选择视频
                            </button>
                        </div>
                    </label>

                    <label className="field-block">
                        <span>输出目录</span>
                        <div className="path-row">
                            <input
                                value={form.outputDir}
                                onChange={(event) =>
                                    setForm((current) => ({
                                        ...current,
                                        outputDir: event.target.value,
                                    }))
                                }
                                placeholder="选择短视频保存位置"
                            />
                            <button type="button" className="secondary-btn" onClick={handleSelectOutputDir}>
                                选择目录
                            </button>
                        </div>
                    </label>

                    <label className="field-block">
                        <span>每段视频时长（秒）</span>
                        <input
                            type="number"
                            min="1"
                            step="1"
                            value={form.segmentLengthSec}
                            onChange={(event) =>
                                setForm((current) => ({
                                    ...current,
                                    segmentLengthSec: event.target.value,
                                }))
                            }
                        />
                    </label>

                    <div className="tips-panel">
                        <div>输出文件命名规则：原文件名 + `_part_001.mp4`</div>
                        <div>适合快速批量拆分长视频素材，直接用于短视频二次处理。</div>
                    </div>

                    <button type="submit" className="primary-btn" disabled={isSubmitting || !environment.ffmpegReady}>
                        {isSubmitting ? "正在切割，请稍候..." : "开始切割"}
                    </button>

                    {errorMessage ? <div className="message error">{errorMessage}</div> : null}

                    {result ? (
                        <div className="message success">
                            <strong>切割完成</strong>
                            <span>共生成 {result.segmentCount} 个短视频</span>
                            <span>保存目录：{result.outputDir}</span>
                        </div>
                    ) : null}
                </form>
            </section>

            <section className="results-layout">
                <article className="panel">
                    <div className="section-heading">
                        <span>处理结果</span>
                        <strong>生成文件</strong>
                    </div>
                    {result?.generatedFiles?.length ? (
                        <div className="file-list">
                            {result.generatedFiles.map((filePath) => (
                                <div className="file-item" key={filePath}>
                                    {filePath}
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="empty-state">切割完成后，这里会显示生成出来的短视频文件列表。</div>
                    )}
                </article>

                <article className="panel">
                    <div className="section-heading">
                        <span>使用说明</span>
                        <strong>最快上手</strong>
                    </div>
                    <div className="instruction-list">
                        <div className="instruction-item">1. 点击“选择视频”，挑选要拆分的原视频。</div>
                        <div className="instruction-item">2. 点击“选择目录”，指定短视频的输出位置。</div>
                        <div className="instruction-item">3. 输入每段时长，例如 `15`、`30`、`60` 秒。</div>
                        <div className="instruction-item">4. 点击“开始切割”，等待生成完成。</div>
                    </div>

                    <div className="ffmpeg-path">
                        <span>当前 ffmpeg 路径</span>
                        <code>{environment.ffmpegPath || "未检测到"}</code>
                    </div>
                </article>
            </section>
        </main>
    );
}

export default App;
