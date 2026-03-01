import { useState, useEffect } from 'react'
import BackgroundAnimation from '../components/BackgroundAnimation'
import PerformanceSection from '../components/PerformanceSection'
import { useScrollAnimation } from '../utils/useScrollAnimation'
import HeroSection from '../components/home/HeroSection'
import FeaturesSection from '../components/home/FeaturesSection'
import QuickStartSection from '../components/home/QuickStartSection'
import APIOverviewSection from '../components/home/APIOverviewSection'
import ComparisonPreviewSection from '../components/home/ComparisonPreviewSection'
import FooterSection from '../components/home/FooterSection'

const Home = () => {
  const isVisible = useScrollAnimation()
  const [starCount, setStarCount] = useState(null)

  useEffect(() => {
    fetch('https://api.github.com/repos/chinmay-sawant/gopdfsuit')
      .then(res => res.json())
      .then(data => {
        if (data.stargazers_count !== undefined) {
          setStarCount(data.stargazers_count)
        }
      })
      .catch(err => console.error('Error fetching stars:', err))
  }, [])

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />
      <HeroSection starCount={starCount} />
      <FeaturesSection isVisible={isVisible} />
      <div className="section-divider container" />
      <QuickStartSection isVisible={isVisible} />
      <div className="section-divider container" />
      <APIOverviewSection isVisible={isVisible} />
      <div className="section-divider container" />
      <section id="section-performance" style={{ padding: '4rem 0' }}>
        <div className="container">
          <PerformanceSection isVisible={isVisible['section-performance']} />
        </div>
      </section>
      <ComparisonPreviewSection isVisible={isVisible} />
      <FooterSection isVisible={isVisible} starCount={starCount} />
    </div>
  )
}

export default Home