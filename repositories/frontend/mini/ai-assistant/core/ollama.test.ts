import { describe, test, expect } from "bun:test";
import type { ChatRequest, ChatResponse, AbortableAsyncIterator } from "ollama";
import { chatRequestWithModel, chatRequestWithThinking, chatWithStream } from "./ollama";
import { chat } from "./ollama";

describe("chatRequestWithModel", () => {
    test("应返回一个函数", () => {
        const option = chatRequestWithModel("llama3.2");
        expect(typeof option).toBe("function");
    });

    test("应将传入的 model 写入到 ChatRequest", () => {
        const req: ChatRequest = { model: "", messages: [] };
        chatRequestWithModel("qwen2.5")(req);
        expect(req.model).toBe("qwen2.5");
    });

    test("应覆盖 ChatRequest 中已存在的 model", () => {
        const req: ChatRequest = { model: "llama3.2", messages: [] };
        chatRequestWithModel("qwen2.5")(req);
        expect(req.model).toBe("qwen2.5");
    });

    test("不应修改 ChatRequest 中的其他字段", () => {
        const messages = [{ role: "user", content: "你好" }];
        const req: ChatRequest = { model: "llama3.2", messages };
        chatRequestWithModel("qwen2.5")(req);
        expect(req.messages).toBe(messages);
        expect(req.messages).toEqual([{ role: "user", content: "你好" }]);
    });

    test("应支持空字符串作为 model", () => {
        const req: ChatRequest = { model: "llama3.2", messages: [] };
        chatRequestWithModel("")(req);
        expect(req.model).toBe("");
    });

    test("多次调用同一个 option 应保持幂等", () => {
        const req: ChatRequest = { model: "", messages: [] };
        const option = chatRequestWithModel("qwen2.5");
        option(req);
        option(req);
        expect(req.model).toBe("qwen2.5");
    });

    test("多个不同 model 的 option 应按调用顺序覆盖", () => {
        const req: ChatRequest = { model: "", messages: [] };
        chatRequestWithModel("llama3.2")(req);
        chatRequestWithModel("qwen2.5")(req);
        chatRequestWithModel("gemma3")(req);
        expect(req.model).toBe("gemma3");
    });

    test("每次调用应返回独立的闭包实例", () => {
        const optionA = chatRequestWithModel("llama3.2");
        const optionB = chatRequestWithModel("qwen2.5");
        expect(optionA).not.toBe(optionB);

        const reqA: ChatRequest = { model: "", messages: [] };
        const reqB: ChatRequest = { model: "", messages: [] };
        optionA(reqA);
        optionB(reqB);
        expect(reqA.model).toBe("llama3.2");
        expect(reqB.model).toBe("qwen2.5");
    });
});

describe("chatRequestWithThinking", () => {
    test("应返回一个函数", () => {
        const option = chatRequestWithThinking(true);
        expect(typeof option).toBe("function");
    });

    test("应将布尔值写入到 ChatRequest.think", () => {
        const req: ChatRequest = { model: "", messages: [] };
        chatRequestWithThinking(true)(req);
        expect(req.think).toBe(true);
    });

    test("应将 thinking 级别写入到 ChatRequest.think", () => {
        const req: ChatRequest = { model: "", messages: [] };
        chatRequestWithThinking("high")(req);
        expect(req.think).toBe("high");
    });

    test("应覆盖 ChatRequest 中已存在的 think", () => {
        const req: ChatRequest = { model: "", messages: [], think: false };
        chatRequestWithThinking("medium")(req);
        expect(req.think).toBe("medium");
    });

    test("不应修改 ChatRequest 中的其他字段", () => {
        const messages = [{ role: "user", content: "你好" }];
        const req: ChatRequest = { model: "llama3.2", messages, stream: true };
        chatRequestWithThinking("low")(req);
        expect(req.model).toBe("llama3.2");
        expect(req.messages).toBe(messages);
        expect(req.stream).toBe(true);
    });

    test("多次调用同一个 option 应保持幂等", () => {
        const req: ChatRequest = { model: "", messages: [] };
        const option = chatRequestWithThinking(true);
        option(req);
        option(req);
        expect(req.think).toBe(true);
    });

    test("多个不同 thinking option 应按调用顺序覆盖", () => {
        const req: ChatRequest = { model: "", messages: [] };
        chatRequestWithThinking(true)(req);
        chatRequestWithThinking("low")(req);
        chatRequestWithThinking("high")(req);
        expect(req.think).toBe("high");
    });
});

describe("chat", () => {
    test("应返回一个 Promise", async () => {
        const result = await chat()
        console.log(result);
    });
});
