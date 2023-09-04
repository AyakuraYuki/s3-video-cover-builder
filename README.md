# s3-video-cover-builder

S3 视频封面、图片缩略图生成器，基于 AWS Lambda 功能实现。

## 部署步骤

1. 终端进入本文件所在目录
2. 执行命令 `make release`
3. 前往 AWS Lambda 控制台，进入左侧菜单列表的 `函数 / Function` 菜单
4. 在右侧函数列表，选择进入 `video-cover-builder-goland` 函数
5. 在 `代码源` 配置，点击 `上传自`，选择 `.zip 文件`
6. 弹出对话框点击 `上传`，定位到本文件所在目录，选择并上传打包好的 `s3-video-cover-builder.zip` 压缩包
7. 等待上传，成功后即部署完成
8. （可选）执行命令 `make cleanup` 清理构件
