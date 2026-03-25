<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>融合平台与智能视频分析系统API接口规格书</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        body {
            font-family: 'Inter', sans-serif;
        }
        /* Custom scrollbar for better aesthetics */
        ::-webkit-scrollbar {
            width: 8px;
        }
        ::-webkit-scrollbar-track {
            background: #f1f1f1;
        }
        ::-webkit-scrollbar-thumb {
            background: #888;
            border-radius: 4px;
        }
        ::-webkit-scrollbar-thumb:hover {
            background: #555;
        }
        .content-section {
            scroll-margin-top: 80px; /* Offset for fixed header */
        }
    </style>
</head>
<body class="bg-gray-50 text-gray-800">

    <div class="flex min-h-screen">
        <!-- Sidebar Navigation -->
        <aside class="w-64 bg-white border-r border-gray-200 p-6 sticky top-0 h-screen overflow-y-auto hidden md:block">
            <h2 class="text-lg font-bold text-gray-900 mb-6">导航目录</h2>
            <nav>
                <ul class="space-y-2">
                    <li><a href="#introduction" class="text-gray-600 hover:text-blue-600 hover:bg-gray-100 block p-2 rounded-md font-medium">1. 概述</a></li>
                    <li><a href="#realtime-alarm" class="text-gray-600 hover:text-blue-600 hover:bg-gray-100 block p-2 rounded-md font-medium">2. 实时告警推送接口</a></li>
                    <li><a href="#realtime-crowd" class="text-gray-600 hover:text-blue-600 hover:bg-gray-100 block p-2 rounded-md font-medium">3. 实时客流数据推送</a></li>
                    <li><a href="#historical-crowd" class="text-gray-600 hover:text-blue-600 hover:bg-gray-100 block p-2 rounded-md font-medium">4. 历史客流数据查询</a></li>
                    <li><a href="#appendix" class="text-gray-600 hover:text-blue-600 hover:bg-gray-100 block p-2 rounded-md font-medium">5. 数据约定</a></li>
                </ul>
            </nav>
        </aside>

        <!-- Main Content -->
        <main class="flex-1 p-6 md:p-10">
            <div class="max-w-4xl mx-auto">
                <header class="mb-10 border-b border-gray-200 pb-6">
                    <h1 class="text-4xl font-bold text-gray-900">融合平台与智能视频分析系统API接口规格书</h1>
                    <p class="text-lg text-gray-500 mt-2">版本号：V 1.5</p>
                </header>

                <!-- Introduction Section -->
                <section id="introduction" class="mb-12 content-section">
                    <h2 class="text-2xl font-semibold text-gray-800 mb-4">1. 概述</h2>
                    <div class="space-y-4 text-gray-700">
                        <p>本文档旨在定义<strong class="font-medium text-gray-900">融合平台</strong>与<strong class="font-medium text-gray-900">智能视频分析系统</strong>之间的API接口标准，作为双方技术团队进行系统集成开发的唯一依据和事实来源。</p>
                        <p>智能视频分析系统负责对指定的摄像头视频码流进行实时AI推理分析。融合平台通过调用本文档定义的接口，获取其关注的分析结果，主要包括<strong class="font-medium text-gray-900">实时告警信息</strong>与<strong class="font-medium text-gray-900">客流统计数据</strong>。</p>
                        <p>系统间的交互模式分为两种：
                            <ul class="list-disc list-inside ml-4 space-y-2">
                                <li><strong>数据推送 (Push)</strong>: 由智能视频分析系统作为客户端，主动将实时告警和实时客流数据推送至融合平台提供的API接口。</li>
                                <li><strong>数据查询 (Pull)</strong>: 由融合平台作为客户端，主动调用智能视频分析系统提供的API接口，查询历史客流等数据。</li>
                            </ul>
                        </p>
                        <p>为确保两个独立系统能够成功联动，除了实现API接口外，双方团队还需共同维护一份统一的数据标准，例如站点名称、摄像头编码和告警类型等。这些标准在本手册的“数据约定”章节中详细定义。</p>
                        <div class="mt-4 text-sm text-gray-600 bg-gray-100 border border-gray-200 rounded-lg p-3">
                            <p><strong class="font-semibold text-gray-800">适用性说明</strong>：本文档中定义的接口和数据约定是为当前项目下的融合平台与智能视频分析系统的特定集成需求而设计的。如需将此文档用于其他项目或平台对接，请注意其存在的局限性，仅可作为参考，相关技术细节需根据实际情况进行重新评估和定义。</p>
                        </div>
                    </div>
                </section>

                <!-- Realtime Alarm Section -->
                <section id="realtime-alarm" class="mb-12 content-section">
                    <h2 class="text-2xl font-semibold text-gray-800 mb-4">2. 实时告警推送接口</h2>
                    
                    <h3 class="text-xl font-medium text-gray-800 mb-3">2.1 概述</h3>
                    <div class="bg-blue-50 border-l-4 border-blue-400 p-4 rounded-r-lg mb-4">
                        <p><strong class="font-semibold">接口类型:</strong> 推送 (Push)</p>
                        <p><strong class="font-semibold">接口提供方:</strong> 融合平台</p>
                        <p><strong class="font-semibold">接口调用方:</strong> 智能视频分析系统</p>
                    </div>
                    <p class="text-gray-700 mb-6">为确保告警信息的可靠传递与最终一致性，智能视频分析系统将周期性调用此接口。每次调用会推送新产生的告警，并重试推送先前失败的历史告警。此机制旨在保证融合平台不会遗漏任何重要的告警事件。</p>
                
                    <h3 class="text-xl font-medium text-gray-800 mb-3">2.2 接口定义</h3>
                    <div class="border border-gray-200 rounded-lg p-4 bg-gray-50">
                        <p><strong class="font-semibold w-24 inline-block">URL</strong>: <code>/CQlineTF/service/S_SP_01</code></p>
                        <p><strong class="font-semibold w-24 inline-block">Method</strong>: <code>POST</code></p>
                        <p><strong class="font-semibold w-24 inline-block">Content-Type</strong>: <code>application/json</code></p>
                        <p><strong class="font-semibold w-24 inline-block">认证</strong>: <span class="bg-yellow-200 text-yellow-800 px-2 py-1 rounded-md text-sm font-medium">无</span></p>
                    </div>
                    <div class="mt-2 text-sm text-gray-600 bg-blue-50 border border-blue-200 rounded-lg p-3">
                        💡 **部署说明**: 融合平台的告警接收服务是<strong class="font-semibold">按站点独立部署</strong>的。因此，智能视频分析系统在推送告警时，必须根据告警所属站点，将请求发送到该站点对应的URL。例如，A站的告警需推送到 <code>http://{A站IP}:{端口}/CQlineTF/service/S_SP_01</code>，B站的告警需推送到 <code>http://{B站IP}:{端口}/CQlineTF/service/S_SP_01</code>。
                    </div>
                
                    <h3 class="text-xl font-medium text-gray-800 mt-6 mb-3">2.3 请求体 (Payload)</h3>
                    <div class="overflow-x-auto">
                        <table class="min-w-full bg-white border border-gray-200 rounded-lg">
                            <thead class="bg-gray-100">
                                <tr>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">参数</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">类型</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">说明</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">alertType</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">告警类型编码。必须与 <a href="#alarm-types" class="text-blue-600 hover:underline">5.2 告警类型编码</a> 中定义的值一致。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">cameraNationalCode</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">20位摄像头国标码。详见 <a href="#camera-code" class="text-blue-600 hover:underline">5.3 摄像头编码</a>。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">hikPlanID</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">5位海康平台规划ID。详见 <a href="#camera-code" class="text-blue-600 hover:underline">5.3 摄像头编码</a>。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">createTime</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">告警事件发生的时间，格式为 <code>YYYY-MM-DD HH:mm:ss</code>。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">file</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">指向告警证据文件（图片或视频）的可访问URL。</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                
                    <h3 class="text-xl font-medium text-gray-800 mt-6 mb-3">2.4 请求示例</h3>
                    <p class="text-gray-700 mb-2">以下示例演示了如何使用 cURL 命令推送一条“隧道积水”告警：</p>
                    <div class="bg-gray-900 text-white p-4 rounded-lg my-4">
                        <pre><code class="language-bash">curl -X POST "http://10.247.84.75:8002/CQlineTF/service/S_SP_01" \
-H "Content-Type: application/json" \
-d '{
    "alertType": "STANDING_WATER",
    "cameraNationalCode": "50000000001310353001",
    "hikPlanID": "10003",
    "createTime": "2025-08-27 14:47:45",
    "file": "http://10.247.84.163:8080/u/A/20250827/001207852804.jpg"
}'</code></pre>
                    </div>
                
                    <h3 class="text-xl font-medium text-gray-800 mt-6 mb-3">2.5 响应说明</h3>
                    <div class="mt-2 text-sm text-gray-600 bg-yellow-50 border border-yellow-200 rounded-lg p-3 mb-4">
                        ⚠️ **重要**: 此接口的成功与否判断比较特殊。无论业务处理是否成功，只要HTTP通信正常，HTTP状态码总是 <code>200 OK</code>。真正的处理结果需要通过解析响应体中的 <code>respCode</code> 字段来确定。
                    </div>
                
                    <h4 class="text-lg font-semibold text-gray-800 mt-4 mb-2">成功响应</h4>
                    <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 mb-4">
                        <p><strong class="font-semibold">条件</strong>: 告警数据被融合平台成功接收并处理。</p>
                        <p><strong class="font-semibold">HTTP 状态码</strong>: <code>200 OK</code></p>
                        <p><strong class="font-semibold">响应体判断</strong>: <code>respCode</code> 的值为 <code>"0"</code>。</p>
                    </div>
                    <div class="bg-gray-900 text-white p-4 rounded-lg my-4">
                        <pre><code class="language-json">{
    "respCode": "0",
    "respDesc": null
}</code></pre>
                    </div>
                
                    <h4 class="text-lg font-semibold text-gray-800 mt-4 mb-2">业务异常响应</h4>
                    <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 mb-4">
                        <p><strong class="font-semibold">条件</strong>: 告警数据格式错误、内容不合法或其它业务逻辑错误。</p>
                        <p><strong class="font-semibold">HTTP 状态码</strong>: <code>200 OK</code></p>
                        <p><strong class="font-semibold">响应体判断</strong>: <code>respCode</code> 的值为非 <code>"0"</code> 的字符串 (例如: <code>"1"</code>)。</p>
                    </div>
                    <div class="bg-gray-900 text-white p-4 rounded-lg my-4">
                        <pre><code class="language-json">{
    "respCode": "1",
    "respDesc": null
}</code></pre>
                    </div>
                    
                    <h4 class="text-lg font-semibold text-gray-800 mt-4 mb-2">通信失败</h4>
                     <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 mb-4">
                        <p><strong class="font-semibold">条件</strong>: 发生网络超时，或收到的 HTTP 状态码不是 <code>200 OK</code>。</p>
                        <p><strong class="font-semibold">处理方式</strong>: 智能视频分析系统应将此次推送标记为失败，并在后续的推送周期中进行重试。</p>
                    </div>
                </section>

                <!-- Realtime Crowd Stats Section -->
                <section id="realtime-crowd" class="mb-12 content-section">
                    <h2 class="text-2xl font-semibold text-gray-800 mb-4">3. 实时客流数据推送接口</h2>
                    
                    <h3 class="text-xl font-medium text-gray-800 mb-3">3.1 概述</h3>
                    <div class="bg-blue-50 border-l-4 border-blue-400 p-4 rounded-r-lg mb-4">
                        <p><strong class="font-semibold">接口类型:</strong> 推送 (Push)</p>
                        <p><strong class="font-semibold">接口提供方:</strong> 融合平台</p>
                        <p><strong class="font-semibold">接口调用方:</strong> 智能视频分析系统</p>
                    </div>
                    <p class="text-gray-700 mb-6">智能视频分析系统根据可配置的周期（默认为1-5分钟）调用此接口，上报客流分析数据。推送内容包含基于AI算法估算的各监控区域内的人数，以及推测的车站总人数。需要注意的是，统计人数为AI推理结果，可能存在一定偏差。</p>
                
                    <h3 class="text-xl font-medium text-gray-800 mb-3">3.2 接口定义</h3>
                    <div class="border border-gray-200 rounded-lg p-4 bg-gray-50">
                        <p><strong class="font-semibold w-24 inline-block">URL</strong>: <code>/CQlineTF/service/S_SP_02</code></p>
                        <p><strong class="font-semibold w-24 inline-block">Method</strong>: <code>POST</code></p>
                        <p><strong class="font-semibold w-24 inline-block">Content-Type</strong>: <code>application/json</code></p>
                        <p><strong class="font-semibold w-24 inline-block">认证</strong>: <span class="bg-yellow-200 text-yellow-800 px-2 py-1 rounded-md text-sm font-medium">无</span></p>
                    </div>
                     <div class="mt-2 text-sm text-gray-600 bg-blue-50 border border-blue-200 rounded-lg p-3">
                        💡 **部署说明**: 与告警接口类似，本接口也建议按站点独立部署，智能视频分析系统需向对应站点的URL推送该站点的客流数据。
                    </div>
                
                    <h3 class="text-xl font-medium text-gray-800 mt-6 mb-3">3.3 请求体 (Payload)</h3>
                    <div class="overflow-x-auto">
                        <table class="min-w-full bg-white border border-gray-200 rounded-lg">
                            <thead class="bg-gray-100">
                                <tr>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">参数</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">类型</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">说明</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">stationName</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">站点名称。其值必须与 <a href="#station-code" class="text-blue-600 hover:underline">5.4 站点位置编码</a> 中定义的官方中文名称一致。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">detectType</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">检测类型编码，例如 "HorizontalCrowdCount"。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">createTime</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">数据生成时间，格式为 <code>YYYY-MM-DD HH:mm:ss</code>。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">crowdCountList</td>
                                    <td class="px-4 py-2 text-sm">Array</td>
                                    <td class="px-4 py-2 text-sm">按摄像头统计的人数列表。每个对象包含 `cameraNationalCode`, `hikPlanID` 和 `crowdCount` (人数)。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">countTypeList</td>
                                    <td class="px-4 py-2 text-sm">Array</td>
                                    <td class="px-4 py-2 text-sm">按区域统计的人数列表。每个对象包含 <code>countType</code> (需使用 <a href="#area-code" class="text-blue-600 hover:underline">5.5</a> 定义的20位编码), <code>countTypeDesc</code> (可选的描述) 和 <code>crowdCount</code> (人数)。</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                
                    <h3 class="text-xl font-medium text-gray-800 mt-6 mb-3">3.4 请求示例</h3>
                    <p class="text-gray-700 mb-2">以下示例演示了如何推送一个站点的实时客流数据：</p>
                    <div class="bg-gray-900 text-white p-4 rounded-lg my-4">
                        <pre><code class="language-bash">curl -X POST "http://10.247.84.75:8002/CQlineTF/service/S_SP_02" \
-H "Content-Type: application/json" \
-d '{
    "stationName": "商贸城站",
    "detectType": "HorizontalCrowdCount",
    "createTime": "2025-08-27 15:09:11",
    "crowdCountList": [
        {"cameraNationalCode": "50000000001310847001", "hikPlanID": "10005", "crowdCount": "30"},
        {"cameraNationalCode": "50000000001310847000", "hikPlanID": "10004", "crowdCount": "70"}
    ],
    "countTypeList": [
        {
            "countType": "50010811240080200101",
            "countTypeDesc": "商贸城站-站厅-区域1",
            "crowdCount": "30"
        },
        {
            "countType": "50010811240080100101",
            "countTypeDesc": "商贸城站-站台-区域1",
            "crowdCount": "70"
        },
        {
            "countType": "50010811240080000101",
            "countTypeDesc": "商贸城站-全站总人数",
            "crowdCount": "100"
        }
    ]
}'</code></pre>
                    </div>
                
                    <h3 class="text-xl font-medium text-gray-800 mt-6 mb-3">3.5 响应说明</h3>
                    <p class="text-gray-700 mb-4">该接口的响应逻辑与实时告警推送接口一致。</p>
                    <h4 class="text-lg font-semibold text-gray-800 mt-4 mb-2">成功响应</h4>
                    <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 mb-4">
                        <p><strong class="font-semibold">HTTP 状态码</strong>: <code>200 OK</code></p>
                        <p><strong class="font-semibold">响应体判断</strong>: <code>respCode</code> 的值为 <code>"0"</code>。</p>
                    </div>
                    <div class="bg-gray-900 text-white p-4 rounded-lg my-4">
                        <pre><code class="language-json">{
    "respCode": "0",
    "respDesc": "NULL"
}</code></pre>
                    </div>
                
                    <h4 class="text-lg font-semibold text-gray-800 mt-4 mb-2">业务异常或通信失败</h4>
                    <p class="text-gray-700">处理方式与 <a href="#realtime-alarm" class="text-blue-600 hover:underline">2.5 响应说明</a> 中的定义完全相同。如果 <code>respCode</code> 非 "0"，或 HTTP 状态码非 200，或发生网络超时，都应视为失败，并在后续周期中重试。</p>

                </section>

                <!-- Historical Crowd Data Section -->
                <section id="historical-crowd" class="mb-12 content-section">
                    <h2 class="text-2xl font-semibold text-gray-800 mb-4">4. 历史客流数据查询接口</h2>
                    
                    <h3 class="text-xl font-medium text-gray-800 mb-3">4.1 概述</h3>
                    <div class="bg-green-50 border-l-4 border-green-400 p-4 rounded-r-lg mb-4">
                        <p><strong class="font-semibold">接口类型:</strong> 查询 (Pull)</p>
                        <p><strong class="font-semibold">接口提供方:</strong> 智能视频分析系统</p>
                        <p><strong class="font-semibold">接口调用方:</strong> 融合平台</p>
                    </div>
                    <p class="text-gray-700 mb-6">该接口采用现代化RESTful风格设计，旨在为融合平台提供一个高效、标准的历史客流数据查询方式，主要用于客流密度热力图等场景。接口返回指定时间范围内，按分析区域/分组聚合后的客流总数。</p>
                
                    <h3 class="text-xl font-medium text-gray-800 mb-3">4.2 接口定义</h3>
                    <div class="border border-gray-200 rounded-lg p-4 bg-gray-50">
                        <p><strong class="font-semibold w-24 inline-block">URL</strong>: <code>/api/v1/analytics/crowd_history</code></p>
                        <p><strong class="font-semibold w-24 inline-block">Method</strong>: <code>POST</code></p>
                        <p><strong class="font-semibold w-24 inline-block">Content-Type</strong>: <code>application/json</code></p>
                        <p><strong class="font-semibold w-24 inline-block">认证</strong>: <span class="bg-yellow-200 text-yellow-800 px-2 py-1 rounded-md text-sm font-medium">无</span></p>
                    </div>
                
                    <h3 class="text-xl font-medium text-gray-800 mt-6 mb-3">4.3 请求体 (Payload)</h3>
                    <p class="text-gray-700 mb-4">查询参数通过JSON格式的请求体传递。所有参数均为可选，可根据需要灵活组合。</p>
                    <div class="overflow-x-auto">
                        <table class="min-w-full bg-white border border-gray-200 rounded-lg">
                            <thead class="bg-gray-100">
                                <tr>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">参数</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">类型</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">说明</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">station_names</td>
                                    <td class="px-4 py-2 text-sm">Array of Strings</td>
                                    <td class="px-4 py-2 text-sm">要查询的站点官方中文名列表。如果省略或为空数组，则查询权限范围内的所有站点。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">start_time</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">查询起始时间，采用 <a href="https://www.iso.org/iso-8601-date-and-time-format.html" target="_blank" class="text-blue-600 hover:underline">ISO 8601</a> 格式的UTC时间字符串 (e.g., <code>"2025-08-27T00:00:00Z"</code>)。如果省略，则默认为30天前。</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">end_time</td>
                                    <td class="px-4 py-2 text-sm">String</td>
                                    <td class="px-4 py-2 text-sm">查询结束时间，格式同上。如果省略，则默认为当前时间。</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                     <div class="mt-2 text-sm text-gray-600 bg-blue-50 border border-blue-200 rounded-lg p-3">
                        💡 **重要提示**：智能视频分析系统仅保留最近30天的数据，超出此范围的查询请求将无法返回结果。
                    </div>
                
                    <h3 class="text-xl font-medium text-gray-800 mt-6 mb-3">4.4 请求示例</h3>
                    <p class="text-gray-700 mb-2">查询“鹿角北站”和“况家塘站”在特定一天内的历史客流数据。</p>
                    <div class="bg-gray-900 text-white p-4 rounded-lg my-4">
                        <pre><code class="language-bash">curl -X POST "http://{分析系统IP}:{端口}/api/v1/analytics/crowd_history" \
-H "Content-Type: application/json" \
-d '{
    "station_names": ["鹿角北站", "况家塘站"],
    "start_time": "2025-08-27T00:00:00Z",
    "end_time": "2025-08-27T23:59:59Z"
}'</code></pre>
                    </div>
                
                    <h3 class="text-xl font-medium text-gray-800 mt-6 mb-3">4.5 响应说明</h3>
                    
                    <h4 class="text-lg font-semibold text-gray-800 mt-4 mb-2">成功响应</h4>
                    <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 mb-4">
                        <p><strong class="font-semibold">HTTP 状态码</strong>: <code>200 OK</code></p>
                        <p class="text-sm">响应体是一个JSON对象，其中 `data` 字段包含了按站点分组的查询结果。每个区域的结果为其在查询时间段内的客流总和。</p>
                    </div>
                    <div class="bg-gray-900 text-white p-4 rounded-lg my-4">
                        <pre><code class="language-json">{
    "response_timestamp": "2025-08-28T08:00:00Z",
    "query_parameters": {
        "station_names": ["鹿角北站", "况家塘站"],
        "start_time": "2025-08-27T00:00:00Z",
        "end_time": "2025-08-27T23:59:59Z"
    },
    "data": [
        {
            "station_name": "鹿角北站",
            "results": [
                {
                    "area_code": "50010811240010100101",
                    "area_description": "鹿角北站-站台-区域1",
                    "total_count": 18532
                },
                {
                    "area_code": "50010811240010200101",
                    "area_description": "鹿角北站-站厅-区域1",
                    "total_count": 9745
                }
            ]
        },
        {
            "station_name": "况家塘站",
            "results": [
                {
                    "area_code": "50010811240020000101",
                    "area_description": "况家塘站-全站总人数",
                    "total_count": 25480
                }
            ]
        }
    ]
}</code></pre>
                    </div>
                
                    <h4 class="text-lg font-semibold text-gray-800 mt-4 mb-2">客户端错误响应</h4>
                    <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 mb-4">
                        <p><strong class="font-semibold">HTTP 状态码</strong>: <code>422 Unprocessable Entity</code></p>
                        <p class="text-sm">当请求体中的参数不符合格式要求时（例如日期格式错误），将返回此状态码。响应体将包含详细的错误信息。</p>
                    </div>
                    <div class="bg-gray-900 text-white p-4 rounded-lg my-4">
                        <pre><code class="language-json">{
    "detail": [
        {
            "loc": [
                "body",
                "start_time"
            ],
            "msg": "invalid datetime format",
            "type": "value_error.datetime"
        }
    ]
}</code></pre>
                    </div>
                    
                    <h4 class="text-lg font-semibold text-gray-800 mt-4 mb-2">服务端错误响应</h4>
                     <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 mb-4">
                        <p><strong class="font-semibold">HTTP 状态码</strong>: <code>500 Internal Server Error</code></p>
                        <p class="text-sm">当视频分析系统内部发生意外错误（如数据库连接失败）导致无法处理请求时，返回此状态码。</p>
                    </div>
                </section>

                <!-- Appendix Section -->
                <section id="appendix" class="content-section">
                    <h2 class="text-2xl font-semibold text-gray-800 mb-4">5. 数据约定</h2>
                    <h3 class="text-xl font-medium text-gray-800 mb-3">5.1 概述</h3>
                    <p class="text-gray-700 mb-4">本章节定义了两个系统间成功通信所必需的共享数据标准和编码规范。为确保数据的一致性、准确性和可解析性，双方开发团队必须严格遵守本章节定义的各项约定。所有在此定义的编码表和数据格式，应被视为系统集成的“唯一真实来源”，任何变更都需经双方同意并同步更新此文档。</p>
                    <div class="mt-2 text-sm text-gray-600 bg-blue-50 border border-blue-200 rounded-lg p-3">
                        <p><strong class="font-semibold text-gray-800">字符编码</strong>：为确保跨平台兼容性（如 Windows, Linux），所有接口中包含非ASCII字符的字符串（例如 `stationName`, `countTypeDesc`, `station_names` 等），都<strong class="font-semibold">必须使用 UTF-8 编码</strong>。</p>
                    </div>
                    
                    <h3 id="alarm-types" class="text-xl font-medium text-gray-800 mt-6 mb-3">5.2 告警类型编码</h3>
                    <p class="text-gray-700 mb-4">为确保告警信息在两个系统间传递时的一致性和可解析性，所有告警类型必须使用下表中定义的统一编码。此列表由双方团队共同维护。</p>
                    <div class="overflow-x-auto">
                        <table class="min-w-full bg-white border border-gray-200 rounded-lg">
                            <thead class="bg-gray-100">
                                <tr>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">功能名称</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">任务类型编码</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">告警类型编码 (alertType)</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr class="border-t">
                                    <td class="px-4 py-2 text-sm">围界闯入</td>
                                    <td class="px-4 py-2 font-mono text-sm">borderDetect</td>
                                    <td class="px-4 py-2 font-mono text-sm">BORDER</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 text-sm">人员摔倒</td>
                                    <td class="px-4 py-2 font-mono text-sm">fallDownDetect</td>
                                    <td class="px-4 py-2 font-mono text-sm">FALLDOWN</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 text-sm">扶梯逆行</td>
                                    <td class="px-4 py-2 font-mono text-sm">reverseDetect</td>
                                    <td class="px-4 py-2 font-mono text-sm">REVERSE</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 text-sm">大体积物品</td>
                                    <td class="px-4 py-2 font-mono text-sm">futiDetect</td>
                                    <td class="px-4 py-2 font-mono text-sm">BIGSIZE_OBJECT</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 text-sm">异常行李</td>
                                    <td class="px-4 py-2 font-mono text-sm">futiDetect</td>
                                    <td class="px-4 py-2 font-mono text-sm">ABORMAL_OBJECT</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 text-sm">人员滞留</td>
                                    <td class="px-4 py-2 font-mono text-sm">LoiteringDetect</td>
                                    <td class="px-4 py-2 font-mono text-sm">LOITERING</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 text-sm">客流密度异常</td>
                                    <td class="px-4 py-2 font-mono text-sm">HorizontalCrowdCount</td>
                                    <td class="px-4 py-2 font-mono text-sm">HEAVY_CROWD</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 text-sm">隧道积水</td>
                                    <td class="px-4 py-2 font-mono text-sm">StandingWaterDetect</td>
                                    <td class="px-4 py-2 font-mono text-sm">STANDING_WATER</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    <h3 id="camera-code" class="text-xl font-medium text-gray-800 mt-6 mb-3">5.3 摄像头编码</h3>
                    <p class="text-gray-700">为兼容融合平台内部不同的系统，API将在适用的情况下同时提供两种摄像头标识符。消费方可根据自身系统需求选择使用。双方系统需共同维护两种编码与物理设备的映射关系。</p>
                    <ul class="list-disc list-inside ml-4 my-4 space-y-2 text-gray-700">
                        <li><strong class="font-semibold text-gray-900">cameraNationalCode</strong>: 20位国标编码 (GB/T 28181)。</li>
                        <li><strong class="font-semibold text-gray-900">hikPlanID</strong>: 5位海康平台规划ID，保证全线唯一。</li>
                    </ul>
                
                    <h3 id="station-code" class="text-xl font-medium text-gray-800 mt-6 mb-3">5.4 站点位置编码</h3>
                    <p class="text-gray-700 mb-4">下表定义了本项目中重庆地铁24号线各站点的统一编号与官方名称。该编号用于构成 <a href="#area-code" class="text-blue-600 hover:underline">5.5 分析区域与分组编码</a> 中的第11-13位。官方名称用于 <a href="#realtime-crowd" class="text-blue-600 hover:underline">3.3 请求体</a> 中的 `stationName` 字段以及 <a href="#historical-crowd" class="text-blue-600 hover:underline">4.3 请求体</a> 中的 `station_names` 字段。</p>
                     <div class="overflow-x-auto">
                        <table class="min-w-full bg-white border border-gray-200 rounded-lg">
                            <thead class="bg-gray-100">
                                <tr>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">站点名称</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">站点编号 (11-13位)</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">鹿角北站</td><td class="px-4 py-2 font-mono text-sm">001</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">况家塘站</td><td class="px-4 py-2 font-mono text-sm">002</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">竹园村站</td><td class="px-4 py-2 font-mono text-sm">003</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">重庆东站</td><td class="px-4 py-2 font-mono text-sm">004</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">地龙湾站</td><td class="px-4 py-2 font-mono text-sm">005</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">瓦子坝站</td><td class="px-4 py-2 font-mono text-sm">006</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">茶涪路站</td><td class="px-4 py-2 font-mono text-sm">007</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">商贸城站</td><td class="px-4 py-2 font-mono text-sm">008</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">迎龙站</td><td class="px-4 py-2 font-mono text-sm">009</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">商贸城北站</td><td class="px-4 py-2 font-mono text-sm">010</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 text-sm">广阳湾站</td><td class="px-4 py-2 font-mono text-sm">011</td></tr>
                            </tbody>
                        </table>
                    </div>

                    <h3 id="area-code" class="text-xl font-medium text-gray-800 mt-6 mb-3">5.5 分析区域与分组编码</h3>
                    <p class="text-gray-700 mb-4">为满足精细化客流统计等AI分析需求，特制定一套结构化、全国唯一的20位编码体系。该编码用于<a href="#realtime-crowd" class="text-blue-600 hover:underline">实时客流数据推送接口</a>中的 `countType` 字段，以及<a href="#historical-crowd" class="text-blue-600 hover:underline">历史客流数据查询接口</a>响应中的 `area_code` 字段。</p>
                    <h4 class="text-lg font-semibold text-gray-800 mt-4 mb-2">编码结构</h4>
                    <div class="overflow-x-auto">
                        <table class="min-w-full bg-white border border-gray-200 rounded-lg">
                            <thead class="bg-gray-100">
                                <tr>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">位码</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">含义</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">长度</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">说明</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr class="border-t"><td class="px-4 py-2 font-mono text-sm">1-6</td><td class="px-4 py-2 text-sm">行政区划代码</td><td class="px-4 py-2 text-sm">6</td><td class="px-4 py-2 text-sm">采用国标 `GB/T 2260`，确保全国唯一。例：重庆市南岸区为 `500108`。</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 font-mono text-sm">7-8</td><td class="px-4 py-2 text-sm">项目类型</td><td class="px-4 py-2 text-sm">2</td><td class="px-4 py-2 text-sm">区分项目。例：`11`=地铁交通, `12`=机场, `13`=智慧园区。</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 font-mono text-sm">9-10</td><td class="px-4 py-2 text-sm">线路/项目编号</td><td class="px-4 py-2 text-sm">2</td><td class="px-4 py-2 text-sm">具体项目编号。例：地铁 `24` 号线。</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 font-mono text-sm">11-13</td><td class="px-4 py-2 text-sm">站点/位置编号</td><td class="px-4 py-2 text-sm">3</td><td class="px-4 py-2 text-sm">线路内的具体位置。例：鹿角北站为 `001`。</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 font-mono text-sm">14-15</td><td class="px-4 py-2 text-sm">主功能区</td><td class="px-4 py-2 text-sm">2</td><td class="px-4 py-2 text-sm">站点内的宏观功能区。例：`00`=全站, `01`=站台, `02`=站厅。</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 font-mono text-sm">16-18</td><td class="px-4 py-2 text-sm">子区域/分组序号</td><td class="px-4 py-2 text-sm">3</td><td class="px-4 py-2 text-sm">主功能区内自定义区域的流水号。</td></tr>
                                <tr class="border-t"><td class="px-4 py-2 font-mono text-sm">19-20</td><td class="px-4 py-2 text-sm">区域类型</td><td class="px-4 py-2 text-sm">2</td><td class="px-4 py-2 text-sm">区分编码性质。例：`01`=物理区域, `02`=逻辑分组。</td></tr>
                            </tbody>
                        </table>
                    </div>
                    <h4 class="text-lg font-semibold text-gray-800 mt-6 mb-2">编码示例</h4>
                    <div class="space-y-4">
                        <div>
                            <p class="font-medium text-gray-800">示例 A: 物理区域</p>
                            <p class="text-gray-700 text-sm mb-2">需求：定义“鹿角北站”站台层西侧的候车区。</p>
                            <div class="bg-gray-900 text-white p-4 rounded-lg">
                                <pre><code>50010811240010100301</code></pre>
                            </div>
                        </div>
                        <div>
                            <p class="font-medium text-gray-800">示例 B: 逻辑分组</p>
                            <p class="text-gray-700 text-sm mb-2">需求：创建一个名为“鹿角北站所有扶梯”的逻辑分组，用于统一分析。</p>
                            <div class="bg-gray-900 text-white p-4 rounded-lg">
                                <pre><code>50010811240010000102</code></pre>
                            </div>
                        </div>
                    </div>

                    <h3 id="area-definition" class="text-xl font-medium text-gray-800 mt-6 mb-3">5.6 分析区域与分组定义</h3>
                    <p class="text-gray-700 mb-4">下表是双方需要共同维护的分析区域/分组定义清单，作为编码与实际业务含义映射的唯一真实来源。所有在接口中使用的编码都必须在此处进行定义。</p>
                    <div class="overflow-x-auto">
                        <table class="min-w-full bg-white border border-gray-200 rounded-lg">
                            <thead class="bg-gray-100">
                                <tr>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">区域/分组编码</th>
                                    <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700">区域/分组信息描述</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">50010811240010000101</td>
                                    <td class="px-4 py-2 text-sm">鹿角北站-全站总人数</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">50010811240010100101</td>
                                    <td class="px-4 py-2 text-sm">鹿角北站-站台-区域1</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">50010811240010200101</td>
                                    <td class="px-4 py-2 text-sm">鹿角北站-站厅-区域1</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">50010811240010100301</td>
                                    <td class="px-4 py-2 text-sm">鹿角北站-站台-西侧候车区</td>
                                </tr>
                                <tr class="border-t">
                                    <td class="px-4 py-2 font-mono text-sm">50010811240010000102</td>
                                    <td class="px-4 py-2 text-sm">鹿角北站-全站所有扶梯 (逻辑分组)</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </section>
            </div>
        </main>
    </div>

</body>
</html>

