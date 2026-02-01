from setuptools import setup, Distribution

class BinaryDistribution(Distribution):
    """Distribution which always forces a binary package with platform name"""
    def has_ext_modules(self):
        return True

setup(
    distclass=BinaryDistribution
)
