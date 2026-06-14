package controller

// assistantHTML is the self-contained chat UI for the finance assistant.
// It calls POST /api/assistant/ask; the browser replays Basic Auth automatically.
const assistantHTML = `<!DOCTYPE html>
<html lang="uk">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Фінансовий асистент</title>
  <style>
    :root { --bg:#0f1115; --panel:#181b22; --me:#2d6cdf; --bot:#22262f; --text:#e6e8eb; --muted:#8b909a; }
    * { box-sizing:border-box; }
    body { margin:0; font-family:-apple-system,Segoe UI,Roboto,sans-serif; background:var(--bg); color:var(--text); height:100vh; display:flex; flex-direction:column; }
    header { padding:16px 20px; border-bottom:1px solid #262a33; font-weight:600; font-size:17px; }
    header span { color:var(--muted); font-weight:400; font-size:13px; margin-left:8px; }
    #chat { flex:1; overflow-y:auto; padding:20px; display:flex; flex-direction:column; gap:12px; max-width:820px; width:100%; margin:0 auto; }
    .msg { padding:12px 16px; border-radius:14px; max-width:80%; line-height:1.5; white-space:pre-wrap; word-wrap:break-word; }
    .me { align-self:flex-end; background:var(--me); border-bottom-right-radius:4px; }
    .bot { align-self:flex-start; background:var(--bot); border-bottom-left-radius:4px; }
    .typing { color:var(--muted); font-style:italic; }
    form { display:flex; gap:10px; padding:16px 20px; border-top:1px solid #262a33; max-width:820px; width:100%; margin:0 auto; }
    input { flex:1; padding:12px 14px; border-radius:10px; border:1px solid #2a2f3a; background:var(--panel); color:var(--text); font-size:15px; outline:none; }
    input:focus { border-color:var(--me); }
    button { padding:12px 20px; border:none; border-radius:10px; background:var(--me); color:#fff; font-size:15px; cursor:pointer; }
    button:disabled { opacity:.5; cursor:default; }
  </style>
</head>
<body>
  <header>💰 Фінансовий асистент <span>запитай про свої витрати</span></header>
  <div id="chat"></div>
  <form id="form">
    <input id="q" placeholder="Скільки я витратив на їжу минулого місяця?" autocomplete="off" autofocus />
    <button id="send" type="submit">Надіслати</button>
  </form>
  <script>
    const chat = document.getElementById('chat');
    const form = document.getElementById('form');
    const input = document.getElementById('q');
    const send = document.getElementById('send');

    function bubble(text, cls) {
      const d = document.createElement('div');
      d.className = 'msg ' + cls;
      d.textContent = text;
      chat.appendChild(d);
      chat.scrollTop = chat.scrollHeight;
      return d;
    }

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      const q = input.value.trim();
      if (!q) return;
      bubble(q, 'me');
      input.value = '';
      send.disabled = true;
      const thinking = bubble('думаю…', 'bot typing');
      try {
        const res = await fetch('/api/assistant/ask', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ question: q }),
        });
        const data = await res.json();
        thinking.remove();
        bubble(res.ok ? data.answer : ('Помилка: ' + (data.error || res.status)), 'bot');
      } catch (err) {
        thinking.remove();
        bubble('Помилка зʼєднання: ' + err.message, 'bot');
      } finally {
        send.disabled = false;
        input.focus();
      }
    });
  </script>
</body>
</html>`
