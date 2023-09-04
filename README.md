# s3-video-cover-builder

S3 视频封面、图片缩略图生成器，基于 AWS Lambda 功能实现。

## 部署步骤

1. 前置条件：需要在 Amazon Lambda 控制台事先创建 FFMpeg 层
2. 终端进入本文件所在目录
3. 执行命令 `make release`
4. 前往 AWS Lambda 控制台，进入左侧菜单列表的 `函数 / Function` 菜单
5. 在右侧函数列表，选择进入 `video-cover-builder-goland` 函数
6. 在 `代码源` 配置，点击 `上传自`，选择 `.zip 文件`
7. 弹出对话框点击 `上传`，定位到本文件所在目录，选择并上传打包好的 `s3-video-cover-builder.zip` 压缩包
8. 等待上传，成功后即部署完成
9. （可选）执行命令 `make cleanup` 清理构件

## 关于 FFMpeg 层

打包上传 FFMpeg 可执行文件压缩包，压缩包内部的结构应该如同下面给出的结构一致：

```text
FFMpeg.zip
.
└── opt
    └── bin
        ├── ffmpeg
        └── ffprobe
```

即：

- `/opt/bin/ffmpeg`
- `/opt/bin/ffprobe`
