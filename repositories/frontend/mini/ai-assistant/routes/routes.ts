import { O } from "ollama/dist/shared/ollama.1bfa89da.mjs";
import { chat } from "../core/ollama";
export const routes = {
    "/api/chat/shift": {
        POST: async (req: Request) => handleChatDuty(req),
    },
    "/api/chat/feedback": {
        POST: async (req: Request) => handleChatFeedback(req),
    },
    "/api/codex/approve": {
        GET: async (req: Request) => handleChatApprove(req),
    },
};

type ShiftClass = "早班" | "晚班";

interface ScheduleContext {
  year: number;
  month: number;
  employees: string[];
  shift_classes: ShiftClass[];
  schedule: Record<string, ShiftClass[]>;
}

interface ChatDutyRequest {
  text: string;
  project_id: string;
  context: ScheduleContext;
}




async function handleChatDuty(req: Request): Promise<Response> {
    const body = (await req.json()) as ChatDutyRequest
    const context = body.context;
    const daysInMonth = new Date(context.year, context.month, 0).getDate();
    const scheduleSummary = Object.entries(context.schedule)
      .map(([name, shifts]) => {
        const summary = (shifts as (string | null)[])
          .map((s, i) => s ? `${i + 1}日:${s}` : null)
          .filter(Boolean)
          .join(', ');
        return `${name}: ${summary || '无排班'}`;
      })
      .join('\n');
    const prompt = `
你是排班调整解析助手。将用户的自然语言描述转为结构化调整列表。
`;
    console.log(prompt)
    const response = await chat([
      { role: "system", content: prompt },
      { role: "user", content: body.text },
    ]);
    return new Response(JSON.stringify({...response}));
}


async function handleChatFeedback(req: Request): Promise<Response> {
    // 1. 接收用户反馈
    // 2. 生成 prompt，调用 Ollama 生成 json 格式的 prd
    // 3. 把反馈和 prd 存到数据库或者日志系统
    // 4. 保存文件到本地
    const body = await req.json();
    console.log("用户反馈:", body);
    // 这里可以把反馈存到数据库或者日志系统
    return new Response(JSON.stringify({ status: "success" }));
} 

async function handleChatPolish(req : Request): Promise<Response> {
  // 1. 接收客服草稿
  // 2. 生成 prompt，调用 Ollama 生成润色后的文本
  // 3. 返回润色后的文本给客服
    const body = await req.json();
    console.log("客服草稿:", body);
    return new Response(JSON.stringify({ status: "success" }));
}

async function handleChatSuggest(req : Request): Promise<Response> {
  // 1. 接收住户和客服的聊天记录
  // 2. 生成 prompt，调用 Ollama 生成改进建议
  // 3. 返回改进建议给客服
    const body = await req.json();
    console.log("住户和客服的聊天记录:", body);
    return new Response(JSON.stringify({ status: "success" }));
}

async function handleChatApprove(req : Request): Promise<Response> {
    // const body = await req.json();

    // console.log(body);
    return new Response(JSON.stringify({ status: "approved" }));
}