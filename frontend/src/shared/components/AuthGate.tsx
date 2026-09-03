import { useEffect, useState, type FormEvent, type ReactNode } from 'react';
import { authApi } from '../api';
import { ArrowRight, BarChart3, Database, LineChart, ShieldCheck } from 'lucide-react';
import { AppModeContext, type AppMode } from '../auth';
import { ThemeToggle } from '../theme';

export function AuthGate({ children }: { children: ReactNode }) {
  const [mode, setMode] = useState<AppMode | null>(null);
  const [username, setUsername] = useState('demo');
  const [password, setPassword] = useState('demo-password');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    authApi.probe().then(() => setMode('LIVE')).catch(() => setMode(null));
    const unauthorized = () => setMode(current => current === 'LIVE' ? null : current);
    window.addEventListener('auth:unauthorized', unauthorized);
    return () => window.removeEventListener('auth:unauthorized', unauthorized);
  }, []);

  async function submit(event: FormEvent, register: boolean) {
    event.preventDefault();
    setBusy(true);
    setError('');
    try {
      if (register) {
        try { await authApi.register(username, password); } catch { /* existing demo user is fine */ }
      }
      await authApi.login(username, password);
      setMode('LIVE');
    } catch {
      setError('Không thể đăng nhập. Kiểm tra backend, database và thông tin tài khoản.');
    } finally {
      setBusy(false);
    }
  }

  if (mode) return <AppModeContext.Provider value={mode}>{children}</AppModeContext.Provider>;
  return <main className="auth-page">
    <ThemeToggle className="auth-theme-toggle" />
    <section className="auth-intro">
      <div className="auth-brand"><BarChart3 size={22} /> Crypto Strategy Lab</div>
      <div>
        <span className="auth-eyebrow">Quantitative research workspace</span>
        <h1>Build strategies.<br />Read the market clearly.</h1>
        <p>One workspace for realtime market data, systematic backtests, strategy discovery and sentiment signals.</p>
      </div>
      <div className="auth-features">
        <span><LineChart size={16} /> Multi-timeframe charts</span>
        <span><Database size={16} /> Verified market history</span>
        <span><ShieldCheck size={16} /> Reproducible backtests</span>
      </div>
    </section>
    <section className="auth-panel">
      <form onSubmit={e=>void submit(e,false)} className="auth-card">
        <div><span className="auth-eyebrow">Secure workspace</span><h2>Welcome back</h2><p>Sign in to connect the API and realtime stream.</p></div>
        <label>Username<input value={username} onChange={e=>setUsername(e.target.value)} minLength={3} placeholder="Username" /></label>
        <label>Password<input value={password} onChange={e=>setPassword(e.target.value)} minLength={8} type="password" placeholder="Password" /></label>
        {error && <small className="auth-error">{error}</small>}
        <button className="auth-primary" type="submit" disabled={busy}>Sign in <ArrowRight size={16} /></button>
        <button className="auth-secondary" type="button" disabled={busy} onClick={e=>void submit(e,true)}>Create demo account</button>
        <div className="auth-divider"><span>or preview without a server</span></div>
        <button className="auth-demo" type="button" disabled={busy} onClick={()=>setMode('DEMO')}>Open offline demo</button>
      </form>
    </section>
  </main>;
}
