import { Link } from "react-router-dom";
import { Mail, Send } from "lucide-react";
import Logo from "./Logo";

export default function Layout({ children }) {
  return (
    <div className="site-shell">
      <header className="header">
        <Logo />
        <nav>
          <Link to="/">Главная</Link>
          <Link to="/category">Категории</Link>
          <a href="#contacts">Контакты</a>
        </nav>
      </header>

      <main>{children}</main>

      <footer id="contacts" className="footer">
        <div>
          <h3>Контакты</h3>

          <div className="contact">
            <Mail size={17} />
            WWchan@gmail.com
          </div>

          <div className="contact">
            <Send size={17} />
            @WWchan
          </div>
        </div>

        <div className="footer-logo">
          <Logo />
        </div>
      </footer>
    </div>
  );
}