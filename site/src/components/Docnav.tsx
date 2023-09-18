import React from 'react'


export default function Docnav() {
  const links = [
    {name: "Home", href: ""},
    {name: "PureMOOtations", href: "rng"},
    {name: "Broadcast", href: "broadcast"},
    {name: "Reaper", href: "reaper"},
    {name: "Managers", href: "manager"},
  ]

  return (
    <>
      <h3 className="mx-2 text-2xl font-bold my-2 text-gld"><a href="/docs">Docs</a></h3>
      <div className="flex-wrap flex">
      {
        links.map(
          (link, i) =>
            <div className={
              "mx-2 my-0 group transition-all text-white"
            } key={i}>
              <a href={"/docs/" + link.href} className="inline-block justify-center font-bold">
                {link.name}
                <span className="block max-w-0 group-focus:max-w-full group-hover:max-w-full transition-all duration-300 h-0.5 bg-pink-600"></span>
              </a>
            </div>
        )
      }
      </div>
    </>
  )
}
