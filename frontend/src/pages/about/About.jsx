import './About.css';

export default function About() {
  return (
    <div className="about-page">
      <section className="about-hero">
        <h1>About FrogMedia</h1>
        <p>
          FrogMedia is a full‑stack project focused on building a scalable,
          real‑world backend‑heavy platform for managing brands, campaigns,
          applications, and real‑time conversations.
        </p>
      </section>

      <section className="about-section">
        <h2>The Project</h2>
        <p>
          The goal of FrogMedia is to simulate the complexity of a modern SaaS
          product — pagination‑heavy feeds, role‑based admin panels, real‑time
          updates, analytics, and clean UX — while remaining production‑oriented
          at every layer.
        </p>
        <p>
          The backend emphasizes strong data modeling, cursor‑based pagination,
          performance‑aware queries, and clear API contracts. The frontend focuses
          on predictable state management, responsiveness, and avoiding UI
          anti‑patterns that don’t scale.
        </p>
      </section>

      <section className="about-section">
        <h2>Tech Philosophy</h2>
        <ul>
          <li>Backend‑first thinking with frontend empathy</li>
          <li>Explicit state and predictable data flow</li>
          <li>Performance before premature abstraction</li>
          <li>Learning by building production‑like systems</li>
        </ul>
      </section>

      <section className="about-section author">
        <h2>The Developer</h2>
        <p>
          Hi, I’m <strong>Sitanshu Mohapatra</strong> — an Electronics
          and Communication Engineering student, <strong>from India</strong>, with a strong interest in backend
          engineering and distributed systems.
        </p>
        <p>
          I primarily work with GoLang, FastAPI, PostgreSQL, MySQL and React, and I enjoy
          designing systems that are simple on the surface but robust underneath.
          FrogMedia is both a learning playground and a serious attempt at imitating
          building software the way it’s done in real products.
        </p>
        <p>
          This project reflects my approach to engineering: understand the
          fundamentals deeply, question abstractions, and optimize for clarity
          and scale.
        </p>
      </section>
    </div>
  );
}

