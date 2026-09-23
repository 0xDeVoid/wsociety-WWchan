import {Link} from "react-router-dom";
import {categories} from "../data/categories";
export default function CategoryGrid(){
  return <div className="category-table" role="table">
  {
  categories.map(cat=><Link className="category-card" to={`/category/${cat.slug}`} key={cat.slug}>
    <div className="cat-top"><span className="cat-code">{cat.code}</span><span className="cat-icon">{cat.icon}</span></div>
    <h3>{cat.title}</h3><p>{cat.text}</p><span className="cat-arrow">→</span>
  </Link>)
  }
</div>}