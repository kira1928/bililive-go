// use proxy middleware function from http-proxy-middleware
const proxy = require('http-proxy-middleware');

module.exports = function(app) {
  // Forward UI static assets from backend
  app.use(
    '/ui',
    proxy({
      target: 'http://127.0.0.1:8080',
      changeOrigin: true,
      // leave /ui path intact for UI static
    })
  );

  // Forward API calls from within UI path to backend API
  app.use(
    '/ui/api',
    proxy({
      target: 'http://127.0.0.1:8080',
      changeOrigin: true,
      pathRewrite: {
        '^/ui/api': '/api',
      },
    })
  );

  // Also proxy standalone /api calls (e.g. local UI to main UI)
  app.use(
    '/api',
    proxy({
      target: 'http://127.0.0.1:8080',
      changeOrigin: true,
    })
  );
};