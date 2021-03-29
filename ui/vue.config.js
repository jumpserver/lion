module.exports = {
    publicPath: '/guacamole/',
    outputDir: 'guacamole',
    assetsDir: 'assets',
    devServer: {
        port: 9528,
        proxy: {
            '^/guacamole': {
                target: 'http://localhost:8081',
                ws: true,
                changeOrigin: true
            }
        }
    }
}