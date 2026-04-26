import ollama from "ollama";
import type { Message } from "ollama";
import type { ChatResponse } from "ollama";
import type { AbortableAsyncIterator } from "ollama";
import type { ChatRequest } from "ollama";


const MODEL_DEFAULT = process.env.OLLAMA_MODEL ?? "minimax-m2.7:cloud";




export type ChatRequestOption = (req: ChatRequest) => void;



export async function chat(
    message: string | Message | Message[] = "hello, world",
    ...options: ChatRequestOption[]
): Promise<ChatResponse> {
    options.push(chatRequestWithStream(false));
    return await internalChat(message, ...options)as ChatResponse;
}

export async function chatWithStream(message: string | Message | Message[] = "hello, world", ...options: ChatRequestOption[]) : Promise<AbortableAsyncIterator<ChatResponse>>{
    options.push(chatRequestWithStream(true));
    return await internalChat(message, ...options)as AbortableAsyncIterator<ChatResponse>;
}

async function internalChat(message: string | Message | Message[], ...options:ChatRequestOption[]): Promise<ChatResponse | AbortableAsyncIterator<ChatResponse>> {
    const req: ChatRequest = {
        model: MODEL_DEFAULT,
        think: false,
    };
    if (typeof message === "string") {
        req.messages = [{ role: "user", content: message }];
    } else if (Array.isArray(message)) {
        req.messages = message;
    } else {
        req.messages = [message];
    }
    for (const option of options) {
        option(req);
    }
    if (req.stream) {
        return ollama.chat(req as ChatRequest & { stream: true });
    } else {
        return ollama.chat(req as ChatRequest & { stream: false });
    }
}


export function chatRequestWithModel(model: string): ChatRequestOption {
    return (req: ChatRequest) => {
        req.model = model;
    };
}

export function chatRequestWithThinking(thinking: boolean | "high" | "medium" | "low"): ChatRequestOption {
    return (req: ChatRequest) => {
        req.think = thinking;
    };
}

export function chatRequestWithStream(stream: boolean): ChatRequestOption {
    return (req: ChatRequest) => {
        req.stream = stream;
    };
}
