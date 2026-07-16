import { fileURLToPath, URL } from 'node:url'

import { defineConfig, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'
import Components from 'unplugin-vue-components/vite'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'

const LION_RUNTIME_ENTRY_MARKER = '<!-- LION_RUNTIME_ENTRY -->'

function createLionRuntimeHtmlPlugin(): Plugin {
  const collectTagAssets = (html: string, pattern: RegExp) => {
    const assets: string[] = []
    const transformedHtml = html.replace(pattern, (_fullMatch, assetPath: string) => {
      assets.push(assetPath)
      return ''
    })

    return {
      assets,
      html: transformedHtml,
    }
  }

  const createRuntimeLoader = (assets: {
    modulePreloads: string[]
    stylesheets: string[]
    moduleScripts: string[]
  }) => {
    return [
      '<script>',
      '  (function () {',
      `    var assets = ${JSON.stringify(assets)};`,
      "    var lionBase = window.__LION_BASE__ || '/';",
      "    var runtimeBase = lionBase.endsWith('/') ? lionBase : lionBase + '/';",
      '',
      '    function resolveAssetPath(assetPath) {',
      '      if (!assetPath) {',
      '        return assetPath;',
      '      }',
      '',
      "      if (/^(https?:)?\\/\\//.test(assetPath)) {",
      '        return assetPath;',
      '      }',
      '',
      "      return runtimeBase + assetPath.replace(/^\\.\\//, '').replace(/^\\//, '');",
      '    }',
      '',
      '    assets.modulePreloads.forEach(function (href) {',
      "      var preloadLink = document.createElement('link');",
      "      preloadLink.rel = 'modulepreload';",
      "      preloadLink.crossOrigin = '';",
      '      preloadLink.href = resolveAssetPath(href);',
      '      document.head.appendChild(preloadLink);',
      '    });',
      '',
      '    assets.stylesheets.forEach(function (href) {',
      "      var stylesheetLink = document.createElement('link');",
      "      stylesheetLink.rel = 'stylesheet';",
      "      stylesheetLink.crossOrigin = '';",
      '      stylesheetLink.href = resolveAssetPath(href);',
      '      document.head.appendChild(stylesheetLink);',
      '    });',
      '',
      '    assets.moduleScripts.forEach(function (src) {',
      "      var moduleScript = document.createElement('script');",
      "      moduleScript.type = 'module';",
      "      moduleScript.crossOrigin = '';",
      '      moduleScript.src = resolveAssetPath(src);',
      '      document.body.appendChild(moduleScript);',
      '    });',
      '  })();',
      '</script>',
    ].join('\n')
  }

  return {
    name: 'lion-runtime-html',
    apply: 'build',
    enforce: 'post',
    transformIndexHtml(html: string) {
      const modulePreloadResult = collectTagAssets(
        html,
        /<link\b[^>]*rel=["']modulepreload["'][^>]*href=["']([^"']+)["'][^>]*>/g,
      )
      const stylesheetResult = collectTagAssets(
        modulePreloadResult.html,
        /<link\b[^>]*rel=["']stylesheet["'][^>]*href=["']([^"']+)["'][^>]*>/g,
      )
      const moduleScriptResult = collectTagAssets(
        stylesheetResult.html,
        /<script\b[^>]*type=["']module["'][^>]*src=["']([^"']+)["'][^>]*><\/script>/g,
      )

      if (
        modulePreloadResult.assets.length === 0 &&
        stylesheetResult.assets.length === 0 &&
        moduleScriptResult.assets.length === 0
      ) {
        return html
      }

      const runtimeLoader = createRuntimeLoader({
        modulePreloads: modulePreloadResult.assets,
        stylesheets: stylesheetResult.assets,
        moduleScripts: moduleScriptResult.assets,
      })

      if (moduleScriptResult.html.includes(LION_RUNTIME_ENTRY_MARKER)) {
        return moduleScriptResult.html.replace(LION_RUNTIME_ENTRY_MARKER, runtimeLoader)
      }

      return moduleScriptResult.html.replace('</body>', `${runtimeLoader}\n  </body>`)
    },
  }
}
// https://vite.dev/config/
export default defineConfig({
  base: './',
  plugins: [
    createLionRuntimeHtmlPlugin(),
    vue(),
    tailwindcss(),
    vueJsx(),
    Components({ dts: true, resolvers: [NaiveUiResolver()] }),
  ],
  resolve: {
    extensions: ['.js', '.ts', '.tsx', '.vue'],
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    port: 9529,
    proxy: {
      '^/lion/ws': {
        target: 'http://localhost:8081',
        ws: true,
        changeOrigin: true,
      },
      '^/lion/api': {
        target: 'http://localhost:8081',
        ws: true,
        changeOrigin: true,
      },
      '^/lion/token': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        ws: true,
      },
    }
  },
})
