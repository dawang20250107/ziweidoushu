# 古籍语料投放规范

后期新增古籍资料、倪海厦著作,按本规范整理为 JSON 即可接入平台,
**无需改代码、无需数据库、支持热加载**。

## 投放方式

二选一:

1. **随二进制打包**:JSON 放入仓库 `data/classics/`,重新构建后随镜像分发;
2. **外部目录热加载**(推荐,内容可独立于代码更新):
   - JSON 放入服务器某目录,设置环境变量 `CORPUS_EXTERNAL_DIR=/path/to/dir`;
   - 启动时自动加载;运行中投放新文件后调用
     `POST /api/v1/admin/corpus/reload`(Bearer ADMIN_TOKEN)即时生效;
   - slug 与内置书目相同则覆盖内置版本。

加载后自动进入全文检索索引,并作为 AI 解读的古籍引文来源。

## Book JSON Schema

一书一文件,UTF-8:

```json
{
  "title": "天机道",
  "slug": "tianjidao",
  "dynasty": "当代",
  "author": "倪海厦",
  "intro": "一句话简介",
  "wordCount": 50000,
  "chapters": [
    {
      "title": "第一章 论紫微",
      "subtitle": "可选副标题",
      "paragraphs": [
        {
          "id": "tjd-1-1",
          "idx": 1,
          "text": "段落原文……",
          "translation": "可选:白话翻译",
          "niNote": "可选:倪师注解"
        }
      ]
    }
  ]
}
```

字段约束:

| 字段 | 必填 | 说明 |
|------|------|------|
| title / slug / chapters | ✓ | slug 全库唯一,仅小写字母数字连字符(用于 URL) |
| paragraphs[].id | ✓ | 全书唯一,建议 `{slug 缩写}-{章}-{段}`(用于锚点定位) |
| paragraphs[].idx | ✓ | 章内序号,1 起 |
| paragraphs[].text | ✓ | 原文(检索与引文以此为准) |
| translation / niNote | | 白话翻译 / 倪师注解,可后补 |

校验:加载时检查必填字段,不合法文件整体拒绝且不影响线上已加载数据。

## 建议切分粒度

- 歌诀类(如骨髓赋):一句/一联为一段;
- 论述类(如全书篇章):一个自然段为一段,长段落 200-500 字为宜
  (检索命中与 AI 引文以段落为单位,过长影响引用精度)。
