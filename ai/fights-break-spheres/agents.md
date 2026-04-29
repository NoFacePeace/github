# Agents Guide

## 目录

chapters: 小说章节目录
characters: 人物风格目录
prompts: 提示词目录
scripts: 分镜脚本目录
images: 图片目录

## 工作流

> 当我开始说某章节的工作时，根据以下步骤执行

1. 从 chapters 找到对应的章节，生成分镜脚本，存放到 scripts 目录
2. 从 scripts 找到对应章节的分镜脚本，生成AI文生图提示词，存放到 prompts 目录
3. 当分镜脚本涉及到人物的时候，从 characters 目录中找到对应人物的提示词
4. 如果 characters 没有对应人物的提示词时，总结该人物的风格，并生成提示词，存放到 characters 目录
5. 如果人物的风格有变化时，总结该人物的风格，并生成新的提示词，追加到 characters 目录下面该人物的文件里面 
6. 从 prompts 找到相应章节的提示词，生成图片，在 images 下面创建章节目录， 存放在下面
7. 如果目录下面已经有相关内容，不需要从头开始，根据已有内容继续工作


