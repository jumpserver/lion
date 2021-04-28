module.exports = {
  publicPath: '/lion/',
  outputDir: 'lion',
  assetsDir: 'assets',
  devServer: {
    port: 9528,
    proxy: {
      '^/lion': {
        target: 'http://localhost:8081',
        ws: true,
        changeOrigin: true,
        bypass: function (req, res, proxyOptions) {
          if (req.headers.accept.indexOf('html') !== -1) {
            console.log('Skipping proxy for browser request.');
            return '/index.html';
          }
        },
      }
    }
  }
}
