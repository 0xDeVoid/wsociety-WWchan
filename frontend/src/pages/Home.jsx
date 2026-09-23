import {Link} from "react-router-dom";
import {ArrowDown, Hash} from "lucide-react";
import CategoryGrid from "../components/CategoryGrid";
import {team} from "../data/team";

export default function Home(){return <>
<section className="hero">
  <div className="hero-art">
    <img src="/assets/mascot.png" alt="WWchan-mascot"/>
    <div className="fake-inputs"><div><Hash size={12}/>WWchan<b>×</b></div><div><Hash size={12}/>/music/<b>×</b></div></div>
  </div>
  <div className="hero-copy">
    <h1>Добро пожаловать в <span>WWchan</span></h1>
    <p>WWchan — анонимная площадка для общения, обсуждений и всего, что приходит в голову. Здесь можно создавать треды, делиться мыслями, мемами, картинками и находить людей с похожими интересами.</p>
    <p className="hero-small">Без лишней регистрации. Без необходимости быть кем-то определённым. Просто общайся.</p>
    <Link className="primary-btn" to="/category">Перейти <ArrowDown size={14}/></Link>
  </div>
</section>
<section className="home-section"><h2>Категории</h2><CategoryGrid/></section>
<section className="home-section team-section"><h2 className="light-title">Команда разработчиков</h2>
  <div className="team-grid">{team.map(([name,image])=><div className="person" key={name}>
    <img src={image} alt={name}/><span>{name}</span>
  </div>)}</div>
</section>
</>}