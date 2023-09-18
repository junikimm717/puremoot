import * as React from 'react'

const links = [
  {name: "Home", href: ""},
  {name: "Docs", href: "docs"},
]


export default function Navbar() {
  return (
    <nav className="flex flex-row justify-start md:justify-between w-full content-center p-4 text-base">
      <div className="font-mono text-pnk hidden md:flex md:items-center">
        <a href="/">
          puremoot.go
        </a>
      </div>
      <div className="font-bold">
        {
          links.map(
            (link, i) =>
              <span className={
                "m-2 group transition-all"
              } key={i}>
                <a href={"/" + link.href} className="inline-block justify-center">
                  {link.name}
                  <span className="block max-w-0 group-focus:max-w-full group-hover:max-w-full transition-all duration-300 h-0.5 bg-pink-600"></span>
                </a>
              </span>
          )
        }

        <span className={ "m-2 group transition-all" }>
          <a href="https://github.com/junikimm717/puremoot" className="inline-block" target="_blank">
            GitHub
            <span className="block max-w-0 group-hover:max-w-full transition-all duration-300 h-0.5 bg-pink-600"></span>
          </a>
        </span>
      </div>
    </nav>
  )
}