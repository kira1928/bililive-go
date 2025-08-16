import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import API from '../utils/api';

const RemoteControl: React.FC = () => {
  // HashRouter puts route and query params into window.location.hash, not search
  const hash = window.location.hash;
  const api = new API();
  const [token, setToken] = useState<string | null>(null);
  const [remoteMode, setRemoteMode] = useState<boolean>(false);
  const [authBase, setAuthBase] = useState<string>('');

  // 解析回调中的 token
  useEffect(() => {
    // 获取远程控制设置及服务器地址
    api.getRemoteSettings().then((rsp: any) => {
      setToken(rsp.token);
      setRemoteMode(rsp.remoteMode);
      let addr = rsp.serverAddr || '';
      if (!/^https?:\/\//.test(addr)) addr = 'http://' + addr;
      setAuthBase(addr);
    }).catch(() => { });

    // 从 hash 中解析 token 查询参数
    const qIndex = hash.indexOf('?');
    const q = qIndex >= 0 ? hash.substring(qIndex + 1) : '';
    const params = new URLSearchParams(q);
    const t = params.get('token');
    if (t) {
      // 保存并更新后端配置
      localStorage.setItem('remote_token', t);
      api.putRemoteSettings(t, true).then(() => {
        api.saveSettingsInBackground();
        setToken(t);
        setRemoteMode(true);
      });
      // 清理 URL
      window.history.replaceState({}, '', '#/remote');
    }
  }, [hash]);

  const handleLogin = () => {
    const redirect = window.location.origin + '#/remote';
    window.location.href = `${authBase}/api/auth/github/login?redirect_uri=${encodeURIComponent(redirect)}`;
  };

  const handleLogout = () => {
    api.putRemoteSettings('', false).then(() => {
      api.saveSettingsInBackground();
      localStorage.removeItem('remote_token');
      setToken(null);
      setRemoteMode(false);
    });
  };

  return (
    <div style={{ padding: '20px' }}>
      <h2>远程控制</h2>
      {remoteMode && token ? (
        <div>
          <p>已登录，Token: <code>{token.substr(0, 8)}…</code></p>
          <button onClick={handleLogout}>退出登录</button>
          <button style={{ marginLeft: '10px' }} onClick={() => {
            api.getRemoteSettings().then((rsp: any) => {
              setToken(rsp.token);
              setRemoteMode(rsp.remoteMode);
            });
          }}>刷新状态</button>
        </div>
      ) : (
        <button onClick={handleLogin} disabled={!authBase}>使用 GitHub 登录启用远程模式</button>
      )}
    </div>
  );
};

export default RemoteControl;